package go_gin_pages

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"tick_test/repository"
	"tick_test/types"
	"tick_test/utils/errDefs"
	"tick_test/utils/jwt"
	"tick_test/utils/random"
	"time"

	"github.com/gin-gonic/gin"
)

type userTokenInfo struct {
	Username string
	Expiry   time.Time
}

var (
	tokenStore      = make(map[string]userTokenInfo)
	tokenStoreMutex sync.RWMutex
)

type accountHandler struct {
	repo repository.AccountRepository
}

func NewAccountHandler(accountRepo repository.AccountRepository) *accountHandler {
	return &accountHandler{repo: accountRepo}
}

func (ah *accountHandler) getPaginatedAccountsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		pageSize, err := strconv.Atoi(c.Query("pageSize"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		pageNumber, err := strconv.Atoi(c.Query("pageNumber"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		if pageNumber <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "pageNumber must be greater than 0"})
			return
		}
		if pageSize <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "pageSize must be greater than 0"})
			return
		}

		accounts, err := ah.repo.FindPaginatedAccounts(pageSize, pageNumber)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, accounts)
	}
}

func (ah *accountHandler) PatchPromoteAccountHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// verify privileges
		claims, err := ah.ConfirmAccountFromGinContext(c)
		if err != nil {
			returnError(c, fmt.Errorf("%w: %v", errDefs.ErrUnauthorized, err))
			return
		}
		if claims.Role != "Admin" {
			returnError(c, fmt.Errorf("%w: only admin can modify roles", errDefs.ErrUnauthorized))
			return
		}

		var data types.AccountPatchPromoteData
		if err := c.ShouldBindJSON(&data); err != nil {
			returnError(c, fmt.Errorf("%w: invalid_json", errDefs.ErrBadRequest))
			return
		}
		if err := ah.repo.PromoteExistingAccount(&data); err != nil {
			returnError(c, err)
			return
		}
		c.JSON(http.StatusOK, nil)
	}
}

func (ah *accountHandler) PatchAccountHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.GetHeader("Username")
		password := c.GetHeader("Password")
		if err := ah.repo.ConfirmAccount(username, password); err != nil {
			returnError(c, err)
			return
		}

		var data types.AccountPatchData
		if err := c.ShouldBindJSON(&data); err != nil {
			returnError(c, fmt.Errorf("%w: %v", errDefs.ErrBadRequest, err.Error()))
			return
		}

		if err := ah.repo.UpdateExistingAccount(username, &data); err != nil {
			returnError(c, err)
			return
		}
		c.JSON(http.StatusOK, nil)
	}
}

func (ah *accountHandler) DeleteAccountHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.GetHeader("Username")

		exists, err := ah.repo.UserExists(username)
		if err != nil {
			returnError(c, err)
			return
		}
		if !exists {
			c.JSON(http.StatusOK, nil)
			return
		}

		password := c.GetHeader("Password")
		if err := ah.repo.ConfirmAccount(username, password); err != nil {
			returnError(c, err)
			return
		}

		if err := ah.repo.DeleteAccount(username); err != nil {
			returnError(c, err)
			return
		}
		c.JSON(http.StatusAccepted, nil)
	}
}

func generateToken(username string) string {
	token := random.RandSeq(80)
	tokenStoreMutex.Lock()
	tokenStore[token] = userTokenInfo{Username: username, Expiry: time.Now().Add(30 * time.Minute)}
	tokenStoreMutex.Unlock()
	return token
}

func confirmToken(val string) (string, error) {
	tokenStoreMutex.RLock()
	info, exists := tokenStore[val]
	tokenStoreMutex.RUnlock()

	if !exists {
		return "", errors.New("invalid token")
	}
	if time.Now().After(info.Expiry) {
		tokenStoreMutex.Lock()
		delete(tokenStore, val)
		tokenStoreMutex.Unlock()
		return "", errors.New("token expired")
	}
	return info.Username, nil
}

func (ah *accountHandler) ConfirmAccountFromGinContext(c *gin.Context) (jwt.Claims, error) {
	if c.GetHeader("Password") != "" {
		username := c.GetHeader("Username")
		password := c.GetHeader("Password")
		err := ah.repo.ConfirmAccount(username, password)
		if err != nil {
			return jwt.Claims{}, fmt.Errorf("%w: %v", errDefs.ErrUnauthorized, "invalid credentials")
		}
		role, _ := ah.repo.FindUserRole(username)
		return jwt.Claims{
			Username: username,
			Role:     role,
		}, err
	}
	if token := c.GetHeader("User-Token"); token != "" {
		claims, err := jwt.ValidateToken(token)
		if err != nil {
			return jwt.Claims{}, fmt.Errorf("%w: %v", errDefs.ErrUnauthorized, err.Error())
		}
		return claims, nil
	}
	return jwt.Claims{}, fmt.Errorf("%w: no token provided", errDefs.ErrUnauthorized)
}

func (ah *accountHandler) LoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.GetHeader("Username")
		password := c.GetHeader("Password")
		if err := ah.repo.ConfirmAccount(username, password); err != nil {
			returnError(c, err)
			return
		}
		role, err := ah.repo.FindUserRole(username)
		if err != nil {
			returnError(c, err)
			return
		}
		token, err := jwt.GenerateToken(username, role, 30*time.Minute)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation error"})
			return
		}
		c.JSON(http.StatusOK, token)

		c.JSON(http.StatusOK, generateToken(username))
	}
}

func (ah *accountHandler) GetAllAccountsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		accounts, err := ah.repo.FindAllAccounts()
		if err != nil {
			returnError(c, err)
			return
		}
		c.JSON(http.StatusOK, accounts)
	}
}

func (ah *accountHandler) PostAccountHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var data types.AccountPostData
		if err := c.ShouldBindJSON(&data); err != nil {
			returnError(c, fmt.Errorf("%w: %v", errDefs.ErrBadRequest, err.Error()))
			return
		}
		if data.Role == "" {
			data.Role = "User"
		}
		if err := ah.repo.SaveAccount(&data); err != nil {
			returnError(c, err)
			return
		}
		c.JSON(http.StatusCreated, data)
	}
}

func (ah *accountHandler) prepareAccount(route *gin.RouterGroup) {
	route.GET("/all", ah.GetAllAccountsHandler())
	route.GET("/", ah.getPaginatedAccountsHandler())
	route.POST("/register", ah.PostAccountHandler())
	route.POST("/login", ah.LoginHandler())
	route.PATCH("/modify", ah.PatchAccountHandler())
	route.PATCH("/promote", ah.PatchPromoteAccountHandler())
	route.DELETE("/delete", ah.DeleteAccountHandler())
}

package go_gin_pages

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"tick_test/repository"
	"tick_test/types"
	"tick_test/utils/errDefs"
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

func NewAccountHandler(accountRepo repository.AccountRepository) (res *accountHandler) {
	return &accountHandler{
		repo: accountRepo,
	}
}

func (ah *accountHandler) PatchPromoteAccountHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// verify privileges
		_, role, err := ah.ConfirmAccountFromGinContext(c)
		if role != "Admin" {
			returnError(c, fmt.Errorf("%w: only admin can modify roles", errDefs.ErrUnauthorized))
			return
		}
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			return
		}

		// apply changes
		var data = new(types.AccountPatchPromoteData)
		if err := c.ShouldBindJSON(data); err != nil {
			returnError(c, fmt.Errorf("%w: invalid_json", errDefs.ErrBadRequest))
			return
		}
		ah.repo.PromoteExistingAccount(data)
		c.JSON(http.StatusOK, nil)
	}
}

func (ah *accountHandler) PatchAccountHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, err := ah.ConfirmUserFromGinContext(c)
		if err != nil {
			returnError(c, err)
			return
		}
		var data = new(types.AccountPatchData)
		if err := c.ShouldBindJSON(data); err != nil {
			returnError(c, fmt.Errorf("%w: %v", errDefs.ErrBadRequest, err.Error()))
			return
		}
		err = ah.repo.UpdateExistingAccount(username, data)
		if err != nil {
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

		username, err = ah.ConfirmUserFromGinContext(c)
		if err != nil {
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

func generateToken(username string) (token string) {
	token = random.RandSeq(80)
	tokenStoreMutex.Lock()
	tokenStore[token] = userTokenInfo{
		Username: username,
		Expiry:   time.Now().Add(30 * time.Minute),
	}
	tokenStoreMutex.Unlock()
	return
}

func confirmToken(val string) (username string, err error) {
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

func (ah *accountHandler) ConfirmUserFromGinContext(c *gin.Context) (username string, err error) {
	if c.GetHeader("Password") != "" {
		username = c.GetHeader("Username")
		password := c.GetHeader("Password")
		err = ah.repo.ConfirmAccount(username, password)
		return
	}
	if token := c.GetHeader("User-Token"); token != "" {
		username, err = confirmToken(token)
		return
	}
	err = fmt.Errorf("%w: can not find suitable verification method", errDefs.ErrUnauthorized)
	return
}

func (ah *accountHandler) ConfirmAccountFromGinContext(c *gin.Context) (username string, role string, err error) {
	username, err = ah.ConfirmUserFromGinContext(c)
	if err != nil {
		return "", "", err
	}

	role, err = ah.repo.FindUserRole(username)
	if err != nil {
		return username, "", fmt.Errorf("error retrieving user role: %w", err)
	}

	return username, role, nil
}

func (ah *accountHandler) LoginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.GetHeader("Username")
		password := c.GetHeader("Password")
		if err := ah.repo.ConfirmAccount(username, password); err != nil {
			returnError(c, err)
			return
		} else {
			token := generateToken(username)
			c.JSON(http.StatusOK, token)
		}
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

		c.JSON(
			http.StatusCreated,
			data,
		)
	}
}

func (ah *accountHandler) prepareAccount(route *gin.RouterGroup) {
	route.GET("/all", ah.repo.EnsureDatabaseIsOK(ah.GetAllAccountsHandler()))
	route.POST("/register", ah.repo.EnsureDatabaseIsOK(ah.PostAccountHandler()))
	route.POST("/login", ah.repo.EnsureDatabaseIsOK(ah.LoginHandler()))
	route.PATCH("/modify", ah.repo.EnsureDatabaseIsOK(ah.PatchAccountHandler()))
	route.PATCH("/promote", ah.repo.EnsureDatabaseIsOK(ah.PatchPromoteAccountHandler()))
	route.DELETE("/delete", ah.repo.EnsureDatabaseIsOK(ah.DeleteAccountHandler()))
}

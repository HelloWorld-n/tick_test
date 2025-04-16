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

func (ah *accountHandler) patchPromoteAccountHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// verify privileges
		_, role, err := ah.confirmAccountFromGinContext(c)
		if role != "Admin" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"Error": fmt.Errorf("%w: only admin can modify roles", errDefs.ErrUnauthorized),
			})
			return
		}
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			return
		}

		// apply changes
		var data = new(types.AccountPatchPromoteData)
		if err := c.ShouldBindJSON(data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		ah.repo.PromoteExistingAccount(data)
		c.JSON(http.StatusOK, nil)
	}
}

func (ah *accountHandler) patchAccountHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, err := ah.confirmUserFromGinContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": err.Error()})
			return
		}
		var data = new(types.AccountPatchData)
		if err := c.ShouldBindJSON(data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		err = ah.repo.UpdateExistingAccount(username, data)
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, nil)
	}
}

func (ah *accountHandler) deleteAccountHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.GetHeader("Username")

		exists, err := ah.repo.UserExists(username)
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			return
		}
		if !exists {
			c.JSON(http.StatusOK, nil)
			return
		}

		username, err = ah.confirmUserFromGinContext(c)
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			return
		}
		if err := ah.repo.DeleteAccount(username); err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
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

func (ah *accountHandler) confirmUserFromGinContext(c *gin.Context) (username string, err error) {
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

func (ah *accountHandler) confirmAccountFromGinContext(c *gin.Context) (username string, role string, err error) {
	username, err = ah.confirmUserFromGinContext(c)
	if err != nil {
		return "", "", err
	}

	role, err = ah.repo.FindUserRole(username)
	if err != nil {
		return username, "", fmt.Errorf("error retrieving user role: %w", err)
	}

	return username, role, nil
}

func (ah *accountHandler) loginHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.GetHeader("Username")
		password := c.GetHeader("Password")
		if err := ah.repo.ConfirmAccount(username, password); err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			return
		} else {
			token := generateToken(username)
			c.JSON(http.StatusOK, token)
		}
	}
}

func (ah *accountHandler) getAllAccountsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		accounts, err := ah.repo.FindAllAccounts()
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, accounts)
	}
}

func (ah *accountHandler) postAccountHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var data types.AccountPostData
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{`Error`: err.Error()})
			return
		}
		if data.Role == "" {
			data.Role = "User"
		}
		if err := ah.repo.SaveAccount(&data); err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, errDefs.ErrDoesExist) {
				status = http.StatusConflict
			}
			c.JSON(status, gin.H{`Error`: err.Error()})
			return
		}

		c.JSON(
			http.StatusCreated,
			data,
		)
	}
}

func (ah *accountHandler) prepareAccount(route *gin.RouterGroup) {
	route.GET("/all", ah.repo.EnsureDatabaseIsOK(ah.getAllAccountsHandler()))
	route.POST("/register", ah.repo.EnsureDatabaseIsOK(ah.postAccountHandler()))
	route.POST("/login", ah.repo.EnsureDatabaseIsOK(ah.loginHandler()))
	route.PATCH("/modify", ah.repo.EnsureDatabaseIsOK(ah.patchAccountHandler()))
	route.PATCH("/promote", ah.repo.EnsureDatabaseIsOK(ah.patchPromoteAccountHandler()))
	route.DELETE("/delete", ah.repo.EnsureDatabaseIsOK(ah.deleteAccountHandler()))
}

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

func patchPromoteAccountHandler(repo *repository.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		// verify privileges
		_, role, err := confirmAccountFromGinContext(c, repo)
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
		repo.PromoteExistingAccount(data)
		c.JSON(http.StatusOK, nil)
	}
}

func patchAccountHandler(repo *repository.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, err := confirmUserFromGinContext(c, repo)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": err.Error()})
			return
		}
		var data = new(types.AccountPatchData)
		if err := c.ShouldBindJSON(data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		err = repo.UpdateExistingAccount(username, data)
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, nil)
	}
}

func deleteAccountHandler(repo *repository.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.GetHeader("Username")

		exists, err := repo.UserExists(username)
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			return
		}
		if !exists {
			c.JSON(http.StatusOK, nil)
			return
		}

		username, err = confirmUserFromGinContext(c, repo)
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			return
		}
		if err := repo.DeleteAccount(username); err != nil {
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

func confirmUserFromGinContext(c *gin.Context, repo *repository.Repo) (username string, err error) {
	if c.GetHeader("Password") != "" {
		username = c.GetHeader("Username")
		password := c.GetHeader("Password")
		err = repo.ConfirmAccount(username, password)
		return
	}
	if token := c.GetHeader("User-Token"); token != "" {
		username, err = confirmToken(token)
		return
	}
	err = fmt.Errorf("%w: can not find suitable verification method", errDefs.ErrUnauthorized)
	return
}

func confirmAccountFromGinContext(c *gin.Context, repo *repository.Repo) (username string, role string, err error) {
	username, err = confirmUserFromGinContext(c, repo)
	if err != nil {
		return "", "", err
	}

	role, err = repo.FindUserRole(username)
	if err != nil {
		return username, "", fmt.Errorf("error retrieving user role: %w", err)
	}

	return username, role, nil
}

func loginHandler(repo *repository.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.GetHeader("Username")
		password := c.GetHeader("Password")
		if err := repo.ConfirmAccount(username, password); err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			return
		} else {
			token := generateToken(username)
			c.JSON(http.StatusOK, token)
		}
	}
}

func getAllAccountsHandler(repo *repository.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		accounts, err := repo.FindAllAccounts()
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, accounts)
	}
}

func postAccountHandler(repo *repository.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data types.AccountPostData
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{`Error`: err.Error()})
			return
		}
		if data.Role == "" {
			data.Role = "User"
		}
		if err := repo.SaveAccount(&data); err != nil {
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

func prepareAccount(route *gin.RouterGroup, repo *repository.Repo) {
	route.GET("/all", repo.EnsureDatabaseIsOK(getAllAccountsHandler(repo)))
	route.POST("/register", repo.EnsureDatabaseIsOK(postAccountHandler(repo)))
	route.POST("/login", repo.EnsureDatabaseIsOK(loginHandler(repo)))
	route.PATCH("/modify", repo.EnsureDatabaseIsOK(patchAccountHandler(repo)))
	route.PATCH("/promote", repo.EnsureDatabaseIsOK(patchPromoteAccountHandler(repo)))
	route.DELETE("/delete", repo.EnsureDatabaseIsOK(deleteAccountHandler(repo)))
}

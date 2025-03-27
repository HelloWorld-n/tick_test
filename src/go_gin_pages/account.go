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

func patchPromoteAccount(c *gin.Context) {
	// verify privileges
	_, role, err := confirmAccountFromGinContext(c)
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
	repository.PromoteExistingAccount(data)
}

func patchAccount(c *gin.Context) {
	username, err := confirmUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": err.Error()})
		return
	}
	var data = new(types.AccountPatchData)
	if err := c.ShouldBindJSON(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	err = repository.UpdateExistingAccount(username, data)
	if err != nil {
		c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nil)
}

func deleteAccount(c *gin.Context) {
	username := c.GetHeader("Username")

	exists, err := repository.UserExists(username)
	if err != nil {
		c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusOK, nil)
		return
	}

	username, err = confirmUserFromGinContext(c)
	if err != nil {
		c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
		return
	}
	if err := repository.DeleteAccount(username); err != nil {
		c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, nil)
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

func confirmUserFromGinContext(c *gin.Context) (username string, err error) {
	if c.GetHeader("Password") != "" {
		username = c.GetHeader("Username")
		password := c.GetHeader("Password")
		err = repository.ConfirmAccount(username, password)
		return
	}
	if token := c.GetHeader("User-Token"); token != "" {
		username, err = confirmToken(token)
		return
	}
	err = fmt.Errorf("%w: can not find suitable verification method", errDefs.ErrUnauthorized)
	return
}

func confirmAccountFromGinContext(c *gin.Context) (username string, role string, err error) {
	username, err = confirmUserFromGinContext(c)
	if err != nil {
		return "", "", err
	}

	role, err = repository.FindUserRole(username)
	if err != nil {
		return username, "", fmt.Errorf("error retrieving user role: %w", err)
	}

	return username, role, nil
}

func login(c *gin.Context) {
	username := c.GetHeader("Username")
	password := c.GetHeader("Password")
	if err := repository.ConfirmAccount(username, password); err != nil {
		c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
		return
	} else {
		token := generateToken(username)
		c.JSON(http.StatusOK, token)
	}
}

func getAllAccounts(c *gin.Context) {
	accounts, err := repository.FindAllAccounts()
	if err != nil {
		c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, accounts)
}

func prepareAccount(route *gin.RouterGroup) {
	route.GET("/all", repository.EnsureDatabaseIsOK(getAllAccounts))
	route.POST("/register", repository.EnsureDatabaseIsOK(repository.CreateAccount))
	route.POST("/login", repository.EnsureDatabaseIsOK(login))
	route.PATCH("/modify", repository.EnsureDatabaseIsOK(patchAccount))
	route.PATCH("/promote", repository.EnsureDatabaseIsOK(patchPromoteAccount))
	route.DELETE("/delete", repository.EnsureDatabaseIsOK(deleteAccount))
}

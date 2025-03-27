package go_gin_pages

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"tick_test/utils/random"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AccountPostData struct {
	Username     string `json:"Username" binding:"gt=4"`
	Password     string `json:"Password" binding:"gt=8"`
	SamePassword string `json:"SamePassword"`
	Role         string `json:"Role"`
}

type AccountPatchData struct {
	Username     string `json:"Username"`
	Password     string `json:"Password"`
	SamePassword string `json:"SamePassword"`
}

type AccountPatchPromoteData struct {
	Username string `json:"Username" binding:"required"`
	Role     string `json:"Role" binding:"required"`
}

type AccountGetData struct {
	Username string `json:"Username" binding:"gt=4"`
	Role     string `json:"Role"`
}

type userTokenInfo struct {
	Username string
	Expiry   time.Time
}

var (
	tokenStore      = make(map[string]userTokenInfo)
	tokenStoreMutex sync.RWMutex
)

func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedBytes), err
}

func confirmPassword(password string, hash string) (err error) {
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return
}

func patchPromoteAccount(c *gin.Context) {
	// verify privileges
	_, role, err := confirmAccountFromGinContext(c)
	if role != "Admin" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"Error": fmt.Errorf("%w: only admin can modify roles", ErrUnauthorized),
		})
		return
	}
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"Error": err,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": err,
			})
		}
		return
	}

	// apply changes
	var data = new(AccountPatchPromoteData)
	if err := c.ShouldBindJSON(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	PromoteExistingAccount(data)
}

func patchAccount(c *gin.Context) {
	username, err := confirmUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": err.Error()})
		return
	}
	var data = new(AccountPatchData)
	if err := c.ShouldBindJSON(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	err = UpdateExistingAccount(username, data)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrBadRequest) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nil)
}

func deleteAccount(c *gin.Context) {
	username := c.GetHeader("Username")

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM account WHERE username = $1);`
	err := database.QueryRow(checkQuery, username).Scan(&exists)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrBadRequest) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"Error": err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusOK, nil)
		return
	}

	username, err = confirmUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": err.Error()})
		return
	}
	_, err = database.Exec(`DELETE FROM account WHERE username = $1`, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
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
		err = ConfirmAccount(username, password)
		return
	}
	if token := c.GetHeader("User-Token"); token != "" {
		username, err = confirmToken(token)
		return
	}
	err = fmt.Errorf("%w: can not find suitable verification method", ErrUnauthorized)
	return
}

func confirmAccountFromGinContext(c *gin.Context) (username string, role string, err error) {
	username, err = confirmUserFromGinContext(c)
	if err != nil {
		return "", "", err
	}

	query := `
		SELECT r.name 
		FROM account a 
		JOIN role r ON a.role_id = r.id 
		WHERE a.username = $1
	`

	err = database.QueryRow(query, username).Scan(&role)
	if err != nil {
		return username, "", fmt.Errorf("error retrieving user role: %w", err)
	}

	return username, role, nil
}

func login(c *gin.Context) {
	username := c.GetHeader("Username")
	password := c.GetHeader("Password")
	if err := ConfirmAccount(username, password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	} else {
		token := generateToken(username)
		c.JSON(http.StatusOK, token)
	}
}

func getAllAccounts(c *gin.Context) {
	accounts, err := FindAllAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, accounts)
}

func prepareAccount(route *gin.RouterGroup) {
	doPostgresPreparationForAccount()

	route.GET("/all", EnsureDatabaseIsOK(getAllAccounts))
	route.POST("/register", EnsureDatabaseIsOK(CreateAccount))
	route.POST("/login", EnsureDatabaseIsOK(login))
	route.PATCH("/modify", EnsureDatabaseIsOK(patchAccount))
	route.PATCH("/promote", EnsureDatabaseIsOK(patchPromoteAccount))
	route.DELETE("/delete", EnsureDatabaseIsOK(deleteAccount))
}

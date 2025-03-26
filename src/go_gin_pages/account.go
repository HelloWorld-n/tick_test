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

func confirmAccount(username string, password string) (err error) {
	query := `SELECT password FROM account WHERE $1 = username`

	rows, err := database.Query(query, username)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		var hash string
		if err = rows.Scan(&hash); err != nil {
			return
		}
		err = confirmPassword(password, hash)
		return
	}

	if err = rows.Err(); err != nil {
		return
	}
	err = errors.New("unable to find user with given username")
	return
}

func createAccount(c *gin.Context) {
	var data AccountPostData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{`Error`: err.Error()})
		return
	}
	if data.Role == "" {
		data.Role = "User"
	}
	if err := saveAccount(&data); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, ErrDoesExist) {
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

func findAllAccounts() (data []AccountGetData, err error) {
	query := `
		SELECT 
			username, 
			(SELECT name FROM role WHERE acc.role_id = id)
		FROM account acc;
	`

	rows, err := database.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	data = make([]AccountGetData, 0)
	for rows.Next() {
		var account AccountGetData
		if err := rows.Scan(&account.Username, &account.Role); err != nil {
			return nil, err
		}
		data = append(data, account)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

func getAllAccounts(c *gin.Context) {
	accounts, err := findAllAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, accounts)
}

func saveAccount(obj *AccountPostData) (err error) {
	if obj.Password != obj.SamePassword {
		return fmt.Errorf("%w: field `Password` differs from field `SamePassword`", ErrBadRequest)
	}

	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM account WHERE username = $1);`
	err = database.QueryRow(checkQuery, obj.Username).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking user existence: %w", err)
	}
	if exists {
		return fmt.Errorf("%w; user with username %s", ErrDoesExist, obj.Username)
	}

	query := `
		INSERT INTO account (
			username, 
			password, 
			role_id
		) VALUES ($1, $2, (SELECT id FROM role WHERE name = $3));
	`
	hashedPassword, err := hashPassword(obj.Password)
	if err != nil {
		return
	}

	_, err = database.Exec(query, obj.Username, hashedPassword, obj.Role)

	return
}

func updateExistingAccount(username string, obj *AccountPatchData) (err error) {
	// verify valid input
	var count int
	err = database.QueryRow(`SELECT COUNT(*) FROM account WHERE username = $1`, username).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("no account found with the specified username")
	}
	if obj.Password != obj.SamePassword {
		return fmt.Errorf("%w: field `Password` differs from field `SamePassword`", ErrBadRequest)
	}
	if obj.Password != "" && len(obj.Password) < 8 {
		return fmt.Errorf("%w: field `Password` is too short; excepted lenght at least 8", ErrBadRequest)
	}

	// apply changes
	if obj.Password != "" {
		hashedPassword, err := hashPassword(obj.Password)
		if err != nil {
			return err
		}
		_, err = database.Exec(`UPDATE account SET password = $1 WHERE username = $2`, hashedPassword, username)
		if err != nil {
			return err
		}
	}
	if obj.Username != "" && obj.Username != username {
		_, err = database.Exec(`UPDATE account SET username = $1 WHERE username = $2`, obj.Username, username)
		if err != nil {
			return err
		}
	}
	return
}

func promoteExistingAccount(obj *AccountPatchPromoteData) (err error) {
	// verify valid input
	var count int
	err = database.QueryRow(`SELECT COUNT(*) FROM account WHERE username = $1`, obj.Username).Scan(&count)
	if err != nil {
		return err
	}

	// apply changes
	if obj.Role != "" {
		_, err = database.Exec(`UPDATE account SET role_id = (SELECT id FROM role WHERE name = $1) WHERE username = $2`, obj.Role, obj.Username)
		if err != nil {
			return err
		}
	}
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
	promoteExistingAccount(data)
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
	err = updateExistingAccount(username, data)
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
		err = confirmAccount(username, password)
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
	if err := confirmAccount(username, password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	} else {
		token := generateToken(username)
		c.JSON(http.StatusOK, token)
	}
}

func doPostgresPreparationForAccount() {
	if database != nil {
		result, err := database.Exec(`
			CREATE TABLE IF NOT EXISTS account (
				username varchar(100) PRIMARY KEY,
				password varchar(500) NOT NULL
			);
		`)
		fmt.Println(result, err)
		result, err = database.Exec(`
			CREATE TABLE IF NOT EXISTS role (
				id SERIAL PRIMARY KEY,
				name TEXT UNIQUE NOT NULL
			);
		`)
		fmt.Println(result, err)
		result, err = database.Exec(`
			INSERT INTO role (name) VALUES
				('User'),
				('BookKeeper'),
				('Admin')
			ON CONFLICT (name) DO NOTHING;
		`)
		fmt.Println(result, err)
		result, err = database.Exec(`
			ALTER TABLE account ADD COLUMN IF NOT EXISTS role_id INT REFERENCES role(id);
		`)
		fmt.Println(result, err)
		result, err = database.Exec(`
			UPDATE account
			SET role_id = (SELECT id FROM role WHERE name = 'User')
			WHERE role_id IS NULL;
		`)
		fmt.Println(result, err)
		result, err = database.Exec(`
			ALTER TABLE account
			ALTER COLUMN role_id SET NOT NULL;
		`)
		fmt.Println(result, err)
	}
}

func prepareAccount(route *gin.RouterGroup) {
	doPostgresPreparationForAccount()

	route.GET("/all", ensureDatabaseIsOK(getAllAccounts))
	route.POST("/register", ensureDatabaseIsOK(createAccount))
	route.POST("/login", ensureDatabaseIsOK(login))
	route.PATCH("/modify", ensureDatabaseIsOK(patchAccount))
	route.PATCH("/promote", ensureDatabaseIsOK(patchPromoteAccount))
	route.DELETE("/delete", ensureDatabaseIsOK(deleteAccount))
}

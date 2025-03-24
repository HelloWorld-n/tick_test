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

type accountPostData struct {
	Username     string `json:"Username" binding:"gt=4"`
	Password     string `json:"Password" binding:"gt=8"`
	SamePassword string `json:"SamePassword"`
	Role         string `json:"Role"`
}

type accountPatchData struct {
	Username     string `json:"Username"`
	Password     string `json:"Password"`
	SamePassword string `json:"SamePassword"`
	Role         string `json:"Role"`
}

type accountGetData struct {
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
		fmt.Println(hash)
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
	var data accountPostData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{`Error`: err.Error()})
		return
	}
	if data.Role == "" {
		data.Role = "User"
	}
	if err := saveAccount(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{`Error`: err.Error()})
		return
	}

	c.JSON(
		http.StatusCreated,
		data,
	)
}

func findAllAccounts() (data []accountGetData, err error) {
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

	data = make([]accountGetData, 0)
	for rows.Next() {
		var account accountGetData
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

func saveAccount(obj *accountPostData) (err error) {
	if obj.Password != obj.SamePassword {
		err = errors.New("field `Password` differs from field `SamePassword`")
		return
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
	fmt.Println(obj.Role)
	_, err = database.Exec(query, obj.Username, hashedPassword, obj.Role)

	fmt.Println(err)
	return
}

func updateExistingAccount(username string, obj *accountPatchData) (err error) {
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
		return errors.New("field `Password` differs from field `SamePassword`")
	}
	if obj.Password != "" && len(obj.Password) < 8 {
		return errors.New("password is too short")
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
	if obj.Role != "" {
		_, err = database.Exec(`UPDATE account SET role_id = (SELECT id FROM role WHERE name = $1) WHERE username = $2`, obj.Role, username)
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

func patchAccount(c *gin.Context) {
	username, err := confirmUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": err.Error()})
		return
	}
	var data = new(accountPatchData)
	if err := c.ShouldBindJSON(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	err = updateExistingAccount(username, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, nil)
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
	err = errors.New("can not find suitable verification method")
	return
}

func login(c *gin.Context) {
	var data accountPostData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if err := confirmAccount(data.Username, data.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	} else {
		token := generateToken(data.Username)
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
}

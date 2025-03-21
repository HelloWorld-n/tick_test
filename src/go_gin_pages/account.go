package go_gin_pages

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type accountData struct {
	Username     string `json:"Username" binding:"gt=4"`
	Password     string `json:"Password" binding:"gt=8"`
	SamePassword string `json:"SamePassword"`
}

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
	if database == nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				`Error`: "database offline",
			},
		)
		return
	}
	var data accountData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{`Error`: err.Error()})
		return
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

func saveAccount(obj *accountData) (err error) {
	if obj.Password != obj.SamePassword {
		err = errors.New("field `Password` differs from field `SamePassword`")
		return
	}

	query := `INSERT INTO account (username, password) VALUES ($1, $2)`
	hashedPassword, err := hashPassword(obj.Password)
	if err != nil {
		return
	}
	_, err = database.Exec(query, obj.Username, hashedPassword)
	return
}

func login(c *gin.Context) {
	if database == nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				`Error`: "database offline",
			},
		)
		return
	}
	var data accountData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if err := confirmAccount(data.Username, data.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	} else {
		// TODO: replace with token that can be used for communication for some time
		c.JSON(http.StatusOK, "OK")
	}
}

func doPostgresPreparationForAccount() {
	if database != nil {
		result, _ := database.Exec(`
			CREATE TABLE IF NOT EXISTS account (
				username varchar(100) PRIMARY KEY,
				password varchar(500) NOT NULL
			);
		`)
		fmt.Println(result)
	}
}

func prepareAccount(route *gin.RouterGroup) {
	doPostgresPreparationForAccount()

	route.POST("/register", createAccount)
	route.POST("/login", login)
}

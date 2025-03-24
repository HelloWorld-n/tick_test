package go_gin_pages

import (
	"fmt"
	"net/http"
	"time"

	"tick_test/types"

	"github.com/gin-gonic/gin"
)

type Message struct {
	From    string            `json:"From"`
	To      string            `json:"To"`
	When    types.ISO8601Date `json:"When"`
	Content string            `json:"Content"`
}

type MessageToSend struct {
	Message Message `json:"Message"`
}

func saveMessage(msg *Message) error {
	query := `INSERT INTO messages (from_user, to_user, content, created_at) VALUES ($1, $2, $3, $4)`
	_, err := database.Exec(query, msg.From, msg.To, msg.Content, msg.When)
	return err
}

func sendMessage(c *gin.Context) {
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "database offline"})
		return
	}
	var data MessageToSend
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	username, err := confirmUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": "user authentication failed; " + err.Error()})
		return
	}

	data.Message.From = username
	data.Message.When = types.ISO8601Date(time.Now().UTC().Format(time.RFC3339))

	if err := saveMessage(&data.Message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, data.Message)
}

func getMessages(c *gin.Context) {
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "database offline"})
		return
	}
	username, err := confirmUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": "user authentication failed; " + err.Error()})
		return
	}

	query := `SELECT from_user, to_user, content, created_at FROM messages WHERE to_user = $1 OR from_user = $1`
	rows, err := database.Query(query, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	defer rows.Close()

	messages := make([]Message, 0)
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.From, &msg.To, &msg.Content, &msg.When); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		messages = append(messages, msg)
	}

	c.JSON(http.StatusOK, messages)
}

func getSentMessages(c *gin.Context) {
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "database offline"})
		return
	}
	username, err := confirmUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": "user authentication failed; " + err.Error()})
		return
	}

	query := `SELECT from_user, to_user, content, created_at FROM messages WHERE from_user = $1`
	rows, err := database.Query(query, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	defer rows.Close()

	messages := make([]Message, 0)
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.From, &msg.To, &msg.Content, &msg.When); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		messages = append(messages, msg)
	}

	c.JSON(http.StatusOK, messages)
}

func getReceivedMessages(c *gin.Context) {
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "database offline"})
		return
	}
	username, err := confirmUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": "user authentication failed; " + err.Error()})
		return
	}

	query := `SELECT from_user, to_user, content, created_at FROM messages WHERE to_user = $1`
	rows, err := database.Query(query, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}
	defer rows.Close()

	messages := make([]Message, 0)
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.From, &msg.To, &msg.Content, &msg.When); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		messages = append(messages, msg)
	}

	c.JSON(http.StatusOK, messages)
}

func doPostgresPreparationForMessages() {
	if database != nil {
		result, err := database.Exec(`
			CREATE TABLE IF NOT EXISTS messages (
				id SERIAL PRIMARY KEY,
				from_user VARCHAR(100) NOT NULL,
				to_user VARCHAR(100) NOT NULL,
				content TEXT NOT NULL, 
				created_at varchar(30) NOT NULL,
				FOREIGN KEY (from_user) REFERENCES account(username),
				FOREIGN KEY (to_user) REFERENCES account(username)
			);
		`)
		fmt.Println(result, err)
	}
}

func prepareMessage(route *gin.RouterGroup) {
	doPostgresPreparationForMessages()

	route.POST("/send", ensureDatabaseIsOK(sendMessage))
	route.GET("/user", ensureDatabaseIsOK(getMessages))
	route.GET("/sent-by", ensureDatabaseIsOK(getSentMessages))
	route.GET("/recv-by", ensureDatabaseIsOK(getReceivedMessages))
}

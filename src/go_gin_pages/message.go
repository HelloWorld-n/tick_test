package go_gin_pages

import (
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

	if err := SaveMessage(&data.Message); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, data.Message)
}
func getMessages(c *gin.Context) {
	username, err := confirmUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": "user authentication failed; " + err.Error()})
		return
	}

	msgs, err := FindMessages(username, true, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, msgs)
}

func getSentMessages(c *gin.Context) {
	username, err := confirmUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": "user authentication failed; " + err.Error()})
		return
	}

	msgs, err := FindMessages(username, true, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, msgs)
}

func getReceivedMessages(c *gin.Context) {
	username, err := confirmUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"Error": "user authentication failed; " + err.Error()})
		return
	}

	msgs, err := FindMessages(username, false, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, msgs)
}

func prepareMessage(route *gin.RouterGroup) {
	route.POST("/send", EnsureDatabaseIsOK(sendMessage))
	route.GET("/user", EnsureDatabaseIsOK(getMessages))
	route.GET("/sent-by", EnsureDatabaseIsOK(getSentMessages))
	route.GET("/recv-by", EnsureDatabaseIsOK(getReceivedMessages))
}

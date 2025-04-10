package go_gin_pages

import (
	"net/http"
	"time"

	"tick_test/repository"
	"tick_test/types"
	"tick_test/utils/errDefs"

	"github.com/gin-gonic/gin"
)

func sendMessage(c *gin.Context) {
	var data types.MessageToSend
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, err := confirmUserFromGinContext(c)
	if err != nil {
		c.JSON(errDefs.DetermineStatus(err), gin.H{"error": "user authentication failed; " + err.Error()})
		return
	}

	data.Message.From = username
	data.Message.When = types.ISO8601Date(time.Now().UTC().Format(time.RFC3339))

	if err := repository.SaveMessage(&data.Message); err != nil {
		c.JSON(errDefs.DetermineStatus(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, data.Message)
}
func getMessages(c *gin.Context) {
	username, err := confirmUserFromGinContext(c)
	if err != nil {
		c.JSON(errDefs.DetermineStatus(err), gin.H{"error": "user authentication failed; " + err.Error()})
		return
	}

	msgs, err := repository.FindMessages(username, true, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, msgs)
}

func getSentMessages(c *gin.Context) {
	username, err := confirmUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user authentication failed; " + err.Error()})
		return
	}

	msgs, err := repository.FindMessages(username, true, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, msgs)
}

func getReceivedMessages(c *gin.Context) {
	username, err := confirmUserFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user authentication failed; " + err.Error()})
		return
	}

	msgs, err := repository.FindMessages(username, false, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, msgs)
}

func prepareMessage(route *gin.RouterGroup) {
	route.POST("/send", repository.EnsureDatabaseIsOK(sendMessage))
	route.GET("/user", repository.EnsureDatabaseIsOK(getMessages))
	route.GET("/sent-by", repository.EnsureDatabaseIsOK(getSentMessages))
	route.GET("/recv-by", repository.EnsureDatabaseIsOK(getReceivedMessages))
}

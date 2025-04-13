package go_gin_pages

import (
	"net/http"
	"time"

	"tick_test/repository"
	"tick_test/types"
	"tick_test/utils/errDefs"

	"github.com/gin-gonic/gin"
)

func sendMessageHandler(repo *repository.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		var data types.MessageToSend
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		username, err := confirmUserFromGinContext(c, repo)
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": "user authentication failed; " + err.Error()})
			return
		}

		data.Message.From = username
		data.Message.When = types.ISO8601Date(time.Now().UTC().Format(time.RFC3339))

		if err := repo.SaveMessage(&data.Message); err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, data.Message)
	}
}

func getMessagesHandler(repo *repository.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, err := confirmUserFromGinContext(c, repo)
		if err != nil {
			c.JSON(errDefs.DetermineStatus(err), gin.H{"Error": "user authentication failed; " + err.Error()})
			return
		}

		msgs, err := repo.FindMessages(username, true, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, msgs)
	}
}

func getSentMessagesHandler(repo *repository.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, err := confirmUserFromGinContext(c, repo)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": "user authentication failed; " + err.Error()})
			return
		}

		msgs, err := repo.FindMessages(username, true, false)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, msgs)
	}
}

func getReceivedMessagesHandler(repo *repository.Repo) gin.HandlerFunc {
	return func(c *gin.Context) {
		username, err := confirmUserFromGinContext(c, repo)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": "user authentication failed; " + err.Error()})
			return
		}

		msgs, err := repo.FindMessages(username, false, true)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, msgs)
	}
}

func prepareMessage(route *gin.RouterGroup, repo *repository.Repo) {
	route.POST("/send", repo.EnsureDatabaseIsOK(sendMessageHandler(repo)))
	route.GET("/user", repo.EnsureDatabaseIsOK(getMessagesHandler(repo)))
	route.GET("/sent-by", repo.EnsureDatabaseIsOK(getSentMessagesHandler(repo)))
	route.GET("/recv-by", repo.EnsureDatabaseIsOK(getReceivedMessagesHandler(repo)))
}

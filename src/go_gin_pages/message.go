package go_gin_pages

import (
	"fmt"
	"net/http"
	"time"

	"tick_test/repository"
	"tick_test/types"
	"tick_test/utils/errDefs"

	"github.com/gin-gonic/gin"
)

type messageHandler struct {
	repo           repository.MessageRepository
	accountHandler *accountHandler
}

func NewMessageHandler(messageRepo repository.MessageRepository) (res *messageHandler) {
	return &messageHandler{
		repo: messageRepo,
	}
}

func (mh *messageHandler) sendMessageHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var data types.MessageToSend
		if err := c.ShouldBindJSON(&data); err != nil {
			returnError(c, fmt.Errorf("%w: %v", errDefs.ErrBadRequest, err.Error()))
			return
		}

		username, err := mh.accountHandler.confirmUserFromGinContext(c)
		if err != nil {
			returnError(c, err)
			return
		}

		data.Message.From = username
		data.Message.When = types.ISO8601Date(time.Now().UTC().Format(time.RFC3339))

		if err := mh.repo.SaveMessage(&data.Message); err != nil {
			returnError(c, err)
			return
		}

		c.JSON(http.StatusCreated, data.Message)
	}
}

func (mh *messageHandler) getMessagesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, err := mh.accountHandler.confirmUserFromGinContext(c)
		if err != nil {
			returnError(c, err)
			return
		}

		msgs, err := mh.repo.FindMessages(username, true, true)
		if err != nil {
			returnError(c, err)
			return
		}

		c.JSON(http.StatusOK, msgs)
	}
}

func (mh *messageHandler) getSentMessagesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, err := mh.accountHandler.confirmUserFromGinContext(c)
		if err != nil {
			returnError(c, err)
			return
		}

		msgs, err := mh.repo.FindMessages(username, true, false)
		if err != nil {
			returnError(c, err)
			return
		}

		c.JSON(http.StatusOK, msgs)
	}
}

func (mh *messageHandler) getReceivedMessagesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, err := mh.accountHandler.confirmUserFromGinContext(c)
		if err != nil {
			returnError(c, err)
			return
		}

		msgs, err := mh.repo.FindMessages(username, false, true)
		if err != nil {
			returnError(c, err)
			return
		}

		c.JSON(http.StatusOK, msgs)
	}
}

func (mh *messageHandler) prepareMessage(route *gin.RouterGroup) {
	route.POST("/send", mh.repo.EnsureDatabaseIsOK(mh.sendMessageHandler()))
	route.GET("/user", mh.repo.EnsureDatabaseIsOK(mh.getMessagesHandler()))
	route.GET("/sent-by", mh.repo.EnsureDatabaseIsOK(mh.getSentMessagesHandler()))
	route.GET("/recv-by", mh.repo.EnsureDatabaseIsOK(mh.getReceivedMessagesHandler()))
}

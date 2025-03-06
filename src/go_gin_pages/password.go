package go_gin_pages

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
)

var passwords []string = make([]string, 0)

type passwordSimpleConfig struct {
	MinSize int      `json:"MinSize"  binding:"required,gt=0"`
	MaxSize int      `json:"MaxSize"  binding:"required,gtefield=MinSize"`
	Charset []string `json:"Charset"  binding:"required,min=2,dive,required"`
}

func findAllPasswords(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		passwords,
	)
}

func createSimplePassword(c *gin.Context) {
	var password string
	var data passwordSimpleConfig
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	for range data.MinSize + rand.Intn(data.MaxSize-data.MinSize+1) {
		randomIndex := rand.Intn(len(data.Charset))
		randomElement := (data.Charset)[randomIndex]
		password = fmt.Sprint(password, randomElement)
	}
	passwords = append(passwords, password)

	c.JSON(
		http.StatusCreated,
		password,
	)
}

func preparePassword(route *gin.RouterGroup) {
	route.GET("", findAllPasswords)
	route.POST("/simple", createSimplePassword)
}

package go_gin_pages

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

var passwords []string = make([]string, 0)

type passwordSimpleConfig struct {
	Size    int      `json:"Size"     binding:"omitempty,gt=0"`
	MinSize int      `json:"MinSize"  binding:"omitempty,gt=0"`
	MaxSize int      `json:"MaxSize"  binding:"omitempty,gtefield=MinSize"`
	Charset []string `json:"Charset"  binding:"required,min=2,dive,required"`
}

func passwordConfigStructLevelValidation(sl validator.StructLevel) {
	config := sl.Current().Interface().(passwordSimpleConfig)

	providedSize := config.Size > 0
	providedMinAndMax := (config.MinSize > 0) && (config.MaxSize > 0)
	providedMinOrMax := (config.MinSize > 0) || (config.MaxSize > 0)
	providedMinXorMan := (config.MinSize > 0) != (config.MaxSize > 0)

	if providedSize && providedMinOrMax {
		sl.ReportError(config.Size, "Size", "Size", "nand", "MinSize,MaxSize")
	}
	if !providedSize && !providedMinAndMax {
		sl.ReportError(config.Size, "Size", "Size", "or", "MinSize,MaxSize")
	}
	if providedMinXorMan {
		sl.ReportError(config.Size, "(MinSize,MaxSize)", "MinSize", "nxor", "MaxSize")
	}
}

func findAllPasswords(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		passwords,
	)
}

func determineSize(data passwordSimpleConfig) int {
	if data.Size > 0 {
		return data.Size
	} else {
		return data.MinSize + rand.Intn(data.MaxSize-data.MinSize+1)
	}
}

func createSimplePasswordFromParsedData(data passwordSimpleConfig) string {
	var password string = ""
	for range determineSize(data) {
		randomIndex := rand.Intn(len(data.Charset))
		randomElement := (data.Charset)[randomIndex]
		password = fmt.Sprint(password, randomElement)
	}
	passwords = append(passwords, password)
	return password
}

func createSimplePassword(c *gin.Context) {
	var data passwordSimpleConfig
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := passwordValidator.Struct(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	password := createSimplePasswordFromParsedData(data)
	c.JSON(
		http.StatusCreated,
		password,
	)
}

func createSimpleStackPassword(c *gin.Context) {
	var data []passwordSimpleConfig
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	password := ""
	for _, item := range data {
		if err := passwordValidator.Struct(item); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		password += createSimplePasswordFromParsedData(item)
	}
	c.JSON(
		http.StatusCreated,
		password,
	)
}

var passwordValidator = validator.New()

func preparePassword(route *gin.RouterGroup) {
	passwordValidator.RegisterStructValidation(passwordConfigStructLevelValidation, passwordSimpleConfig{})

	route.GET("", findAllPasswords)
	route.POST("/simple", createSimplePassword)
	route.POST("/simple-stack", createSimpleStackPassword)
}

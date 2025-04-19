package go_gin_pages

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

var passwords []string = make([]string, 0)

type runeChecker struct {
	Score   int
	Checker func(rune) bool
}

var runeCheckers = make([]runeChecker, 0)

type PasswordSimpleConfig struct {
	Size    int      `json:"size"     binding:"omitempty,gt=0"`
	MinSize int      `json:"minSize"  binding:"omitempty,gt=0"`
	MaxSize int      `json:"maxSize"  binding:"omitempty,gtefield=MinSize"`
	Charset []string `json:"charset"  binding:"required,min=2,dive,required"`
}

type PasswordSimpleStackConfig []struct {
	PasswordSimpleConfig
	InclusionChances float64 `json:"inclusionChances" binding:"omitempty,min=0,max=1"`
}

func (conf PasswordSimpleStackConfig) extractPasswordSimpleConfig(i int) PasswordSimpleConfig {
	return PasswordSimpleConfig{
		Size:    conf[i].Size,
		MinSize: conf[i].MinSize,
		MaxSize: conf[i].MaxSize,
		Charset: conf[i].Charset,
	}
}

func passwordConfigStructLevelValidation(sl validator.StructLevel) {
	config := sl.Current().Interface().(PasswordSimpleConfig)

	providedSize := config.Size > 0
	providedMinAndMax := (config.MinSize > 0) && (config.MaxSize > 0)
	providedMinOrMax := (config.MinSize > 0) || (config.MaxSize > 0)
	providedMinXorMan := (config.MinSize > 0) != (config.MaxSize > 0)

	if providedSize && providedMinOrMax {
		sl.ReportError(config.Size, "size", "size", "nand", "minSize,maxSize")
	}
	if !providedSize && !providedMinAndMax {
		sl.ReportError(config.Size, "size", "size", "or", "minSize,maxSize")
	}
	if providedMinXorMan {
		sl.ReportError(config.Size, "(minSize,maxSize)", "minSize", "nxor", "maxSize")
	}
}

func findAllPasswords(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		passwords,
	)
}

func ratePassword(c *gin.Context) {
	password := c.Param("password")
	score := 0
	passwordLen := len(password)
	for _, pLen := range []int{5, 8, 14, 20} {
		if passwordLen > pLen {
			score += 1
		} else {
			break
		}
	}
	for _, glyphChecker := range runeCheckers {
		if strings.IndexFunc(password, glyphChecker.Checker) != -1 {
			score += glyphChecker.Score
		}
	}
	c.JSON(
		http.StatusOK,
		gin.H{
			"password": password,
			"score":    score,
		},
	)
}

func determineSize(data PasswordSimpleConfig) int {
	if data.Size > 0 {
		return data.Size
	} else {
		return data.MinSize + rand.Intn(data.MaxSize-data.MinSize+1)
	}
}

func createSimplePasswordFromParsedData(data PasswordSimpleConfig) string {
	var password string = ""
	for range determineSize(data) {
		randomIndex := rand.Intn(len(data.Charset))
		randomElement := (data.Charset)[randomIndex]
		password = fmt.Sprint(password, randomElement)
	}
	return password
}

func createSimplePassword(c *gin.Context) {
	var data PasswordSimpleConfig
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := passwordValidator.Struct(data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	password := createSimplePasswordFromParsedData(data)
	passwords = append(passwords, password)
	c.JSON(
		http.StatusCreated,
		password,
	)
}

func createSimpleStackPassword(c *gin.Context) {
	var data PasswordSimpleStackConfig
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	password := ""
	for i, item := range data {
		if err := passwordValidator.Struct(item); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if item.InclusionChances == 0 || rand.Float64() < item.InclusionChances {
			password += createSimplePasswordFromParsedData(data.extractPasswordSimpleConfig(i))
		}
	}
	passwords = append(passwords, password)
	c.JSON(
		http.StatusCreated,
		password,
	)
}

var passwordValidator = validator.New()

func preparePassword(route *gin.RouterGroup) {
	passwordValidator.RegisterStructValidation(passwordConfigStructLevelValidation, PasswordSimpleConfig{})

	runeCheckers = append(runeCheckers, runeChecker{
		Score:   3,
		Checker: func(glyph rune) bool { return glyph > unicode.MaxASCII },
	})
	runeCheckers = append(runeCheckers, runeChecker{
		Score:   1,
		Checker: unicode.IsUpper,
	})
	runeCheckers = append(runeCheckers, runeChecker{
		Score:   1,
		Checker: unicode.IsLower,
	})
	runeCheckers = append(runeCheckers, runeChecker{
		Score:   1,
		Checker: unicode.IsSymbol,
	})

	route.GET("", findAllPasswords)
	route.GET("/rate/:password", ratePassword)
	route.POST("/simple", createSimplePassword)
	route.POST("/simple-stack", createSimpleStackPassword)
}

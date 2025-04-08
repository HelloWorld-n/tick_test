package go_gin_pages

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"tick_test/repository"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

const urlFile = "../.config/url.txt"
const configFile = "../config.yaml"

type Config struct {
	BaseURL string `yaml:"BaseUrl"`
	Port    int    `yaml:"Port"`
}

func index(c *gin.Context) {
	if err := repository.LoadIteration(); err != nil {
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"Error": fmt.Sprintf("Failed to load iteration: %v", err),
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		repository.ResultIndex{
			Now:       time.Now().UTC().Format(time.RFC3339),
			Iteration: repository.Iteration,
		},
	)
}

type corsMiddleware struct {
	origin string
}

func DetermineURL() (url string, err error) {
	url = "127.0.0.1:4041"
	file, err := os.Open(urlFile)
	if err != nil {
		return
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	if err == nil {
		url = strings.TrimSpace(string(b))
	}
	return
}

func UseConfigToDetermineURL() (url string, err error) {
	config := Config{
		BaseURL: "localhost",
		Port:    4041,
	}
	url = fmt.Sprint(config.BaseURL, ":", config.Port)

	file, err := os.Open(configFile)
	if err != nil {
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return
	}

	if err = yaml.Unmarshal(data, &config); err != nil {
		return
	}

	url = fmt.Sprint(config.BaseURL, ":", config.Port)

	return
}

func Prepare(engine *gin.Engine, url string) {
	cmw := &corsMiddleware{
		origin: url,
	}

	engine.Use(func(c *gin.Context) {
		allowedOrigins := map[string]bool{
			"http://" + cmw.origin:  true,
			"https://" + cmw.origin: true,
			"ws://" + cmw.origin:    true,
			"wss://" + cmw.origin:   true,
		}

		requestOrigin := c.Request.Header.Get("Origin")
		if allowedOrigins[requestOrigin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", requestOrigin)
			c.Writer.Header().Set("Vary", "Origin")
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "null")
		}

		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Header().Set("Content-Security-Policy", "connect-src 'self' "+requestOrigin)

		if c.Request.Method == http.MethodOptions {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "false")
			c.Writer.Header().Set("Access-Control-Expose-Headers", "*")
			c.Writer.Header().Set("Access-Control-Max-Age", "900")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, Authorization, X-Requested-With, Username, Password, User-Token")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	repository.DoPostgresPreparation()
	engine.GET("/", index)
	prepareManipulator(engine.Group("/manipulator"))
	prepareSort(engine.Group("/sort"))
	preparePassword(engine.Group("/password"))
	prepareAccount(engine.Group("/account"))
	prepareMessage(engine.Group("/message"))
	prepareBook(engine.Group("/book"))
}

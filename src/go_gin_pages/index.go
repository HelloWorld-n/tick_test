package go_gin_pages

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"tick_test/types"

	"github.com/gin-gonic/gin"
)

type resultIndex struct {
	Iteration int               `json:"Iteration"`
	Now       types.ISO8601Date `json:"Now"`
}

var iteration int

const iterationFile = "../.data/Iteration.json"
const dbPathFile = "../.config/dbPath.txt"
const urlFile = "../.congih/url.txt"

var iterationMutex sync.Mutex

var ErrDatabaseOffline = errors.New("database offline")
var ErrDoesExist = errors.New("item already exists")
var ErrBadRequest = errors.New("bad request")
var ErrMissingField = fmt.Errorf("%w: field missing", ErrBadRequest)
var ErrUnauthorized = errors.New("unauthorized")

func saveIteration() error {
	if err := os.MkdirAll(filepath.Dir(iterationFile), 0755); err != nil {
		return err
	}

	file, err := os.Create(iterationFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(iteration); err != nil {
		return err
	}
	return nil
}

func index(c *gin.Context) {
	if err := loadIteration(); err != nil {
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
		resultIndex{
			Now:       time.Now().UTC().Format(time.RFC3339),
			Iteration: iteration,
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
		url = string(b)
	}
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

	DoPostgresPreparation()
	loadIteration()
	engine.GET("/", index)
	prepareManipulator(engine.Group("/manipulator"))
	prepareSort(engine.Group("/sort"))
	preparePassword(engine.Group("/password"))
	prepareAccount(engine.Group("/account"))
	prepareMessage(engine.Group("/message"))
	prepareBook(engine.Group("/book"))
}

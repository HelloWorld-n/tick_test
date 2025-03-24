package go_gin_pages

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tick_test/sql_conn"
	"tick_test/types"

	"github.com/gin-gonic/gin"
)

type resultIndex struct {
	Iteration int               `json:"Iteration"`
	Now       types.ISO8601Date `json:"Now"`
}

var iteration int
var database *sql.DB

const iterationFile = "../.data/Iteration.json"
const dbPathFile = "../.config/dbPath.txt"
const urlFile = "../.congih/url.txt"

var iterationMutex sync.Mutex

var ErrDatabaseOffline = errors.New("database offline")
var ErrDoesExist = errors.New("item already exists")

func ensureDatabaseIsOK(fn func(*gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		if database == nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					`Error`: ErrDatabaseOffline,
				},
			)
			return
		}
		fn(c)
	}
}

func loadDatabasePath() (url string, err error) {
	url = ""
	file, err := os.Open(dbPathFile)
	if err != nil {
		return
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	url = string(b)

	return
}

func loadIteration() error {
	iterationMutex.Lock()
	defer iterationMutex.Unlock()

	file, err := os.Open(iterationFile)
	if err != nil {
		if os.IsNotExist(err) {
			iteration = 0
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&iteration); err != nil {
		return err
	}
	return nil
}

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

func doPostgresPreparation() {
	databasePath, err := loadDatabasePath()
	if err != nil {
		return
	}
	databasePath = strings.TrimSpace(databasePath)
	db, err := sql_conn.Prepare(databasePath)
	if err != nil {
		fmt.Println(err)
	} else {
		database = db
	}
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

	doPostgresPreparation()
	loadIteration()
	engine.GET("/", index)
	prepareManipulator(engine.Group("/manipulator"))
	prepareSort(engine.Group("/sort"))
	preparePassword(engine.Group("/password"))
	prepareAccount(engine.Group("/account"))
	prepareMessage(engine.Group("/message"))
}

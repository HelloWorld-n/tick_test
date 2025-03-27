package go_gin_pages

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"tick_test/sql_conn"

	"github.com/gin-gonic/gin"
)

var database *sql.DB

func DoPostgresPreparation() {
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

func EnsureDatabaseIsOK(fn func(*gin.Context)) func(c *gin.Context) {
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

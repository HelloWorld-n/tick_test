package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"tick_test/sql_conn"
	"tick_test/types"
	"tick_test/utils/errDefs"

	"github.com/gin-gonic/gin"
)

const iterationFile = "../.data/Iteration.json"
const dbPathFile = "../.config/dbPath.txt"

type ResultIndex struct {
	Iteration int               `json:"Iteration"`
	Now       types.ISO8601Date `json:"Now"`
}

var Iteration int
var IterationMutex sync.Mutex

func (r *repo) IsDatabaseEnabled() bool {
	return r.DB.Conn != nil
}

func (r *repo) DoPostgresPreparation() (db *sql.DB, err error) {
	databasePath, err := LoadDatabasePath()
	if err != nil {
		return
	}
	db, err = sql_conn.Prepare(databasePath)

	if err != nil {
		fmt.Println(err)
		return
	} else {
		r.DB.Conn = db
	}

	r.doPostgresPreparationForMessages()
	r.doPostgresPreparationForAccount()
	r.doPostgresPreparationForBook()
	r.doPostgresPreparationForManipulator()
	r.loadIterationManipulators()
	LoadIteration()
	return
}

func (r *repo) EnsureDatabaseIsOK(fn func(*gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		if r.DB.Conn == nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					`Error`: errDefs.ErrDatabaseOffline,
				},
			)
			return
		}
		fn(c)
	}
}

func LoadDatabasePath() (url string, err error) {
	url = ""
	file, err := os.Open(dbPathFile)
	if err != nil {
		return
	}
	defer file.Close()

	b, err := io.ReadAll(file)
	url = string(b)
	url = strings.TrimSpace(url)
	return
}

func LoadIteration() error {
	IterationMutex.Lock()
	defer IterationMutex.Unlock()

	file, err := os.Open(iterationFile)
	if err != nil {
		if os.IsNotExist(err) {
			Iteration = 0
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&Iteration); err != nil {
		return err
	}
	return nil
}

func SaveIteration() error {
	if err := os.MkdirAll(filepath.Dir(iterationFile), 0755); err != nil {
		return err
	}

	file, err := os.Create(iterationFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(Iteration); err != nil {
		return err
	}
	return nil
}

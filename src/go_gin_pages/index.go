package go_gin_pages

import (
	"encoding/json"
	"fmt"
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

var iterationMutex sync.Mutex

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

func Prepare(engine *gin.Engine) {
	loadIteration()
	engine.GET("/", index)
	prepareManipulator(engine.Group("/manipulator"))
	prepareSort(engine.Group("/sort"))
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

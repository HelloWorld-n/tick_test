package go_gin_pages

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"tick_test/types"
	"tick_test/utils/random"

	"github.com/gin-gonic/gin"
	"github.com/kodergarten/iso8601duration"
)

type manipulateIterationData struct {
	Duration types.ISO8601Duration `json:"Duration" binding:"required"`
	Value    int                   `json:"Value" binding:"required"`
}

type updateIterationManipulatorData struct {
	Duration *types.ISO8601Duration `json:"Duration"`
	Value    *int                   `json:"Value"`
}

type iterationManipulator struct {
	Code        string                  `json:"Code"`
	Data        manipulateIterationData `json:"Data"`
	Manipulator *time.Ticker            `json:"-"`
}

const iterationManipulatorFile = "../.data/IterationManipulators.json"

var iterationManipulatorMutex sync.Mutex

func loadIterationManipulatorsFromFile() error {
	file, err := os.Open(iterationManipulatorFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&iterationManipulators); err != nil {
		return err
	}
	return nil
}

func loadIterationManipulators() error {
	if database != nil {
		if err := loadIterationManipulatorsFromDatabase(); err != nil {
			return err
		}
	} else {
		if err := loadIterationManipulatorsFromFile(); err != nil {
			return err
		}
	}
	for _, iterationManipulator := range iterationManipulators {
		dur, err := parseISO8601Duration(iterationManipulator.Data.Duration, time.Second)
		if err != nil {
			return err
		}
		ticker := time.NewTicker(dur)
		iterationManipulator.Manipulator = ticker
		go manipulateIteration(iterationManipulator)
	}
	return nil
}

func saveIterationManipulators() error {
	iterationManipulatorMutex.Lock()
	defer iterationManipulatorMutex.Unlock()
	if err := os.MkdirAll(filepath.Dir(iterationManipulatorFile), 0755); err != nil {
		return err
	}

	file, err := os.Create(iterationManipulatorFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(iterationManipulators); err != nil {
		return err
	}
	return nil
}

var iterationManipulators []*iterationManipulator = make([]*iterationManipulator, 0)

func manipulateIteration(obj *iterationManipulator) error {
	for range obj.Manipulator.C {
		iterationMutex.Lock()
		iteration += obj.Data.Value
		if err := saveIteration(); err != nil {
			fmt.Printf("Error saving iteration: %v\n", err)
		}
		iterationMutex.Unlock()
	}
	return nil
}

func parseISO8601Duration(val types.ISO8601Duration, minDuration time.Duration) (dur time.Duration, err error) {
	duration, err := iso8601duration.ParseString(val)
	if err != nil {
		return
	}
	dur = duration.ToDuration()
	if dur < minDuration {
		err = errors.New("field Duration needs to be higher")
		return
	}
	return
}

func findAllIterationManipulators(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		iterationManipulators,
	)
}

func findIterationManipulatorByCode(c *gin.Context) {
	code := c.Param("code")

	for _, v := range iterationManipulators {
		if v.Code == code {
			c.JSON(
				http.StatusOK,
				v.Data,
			)
			return
		}
	}
	c.Status(http.StatusNoContent)
}

func createIterationManipulator(c *gin.Context) {
	if database == nil {
		defer saveIterationManipulators()
	}
	var data manipulateIterationData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dur, err := parseISO8601Duration(data.Duration, time.Second)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ticker := time.NewTicker(dur)

	iterationManipulator := iterationManipulator{
		Code:        random.RandSeq(80),
		Data:        data,
		Manipulator: ticker,
	}
	ensureUniqueCodeForIterationManipulator(&iterationManipulator)
	err = saveIterationManipulatorToDatabase(&iterationManipulator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	iterationManipulators = append(iterationManipulators, &iterationManipulator)
	go manipulateIteration(&iterationManipulator)
	c.JSON(
		http.StatusCreated,
		iterationManipulator,
	)
}

func ensureUniqueCodeForIterationManipulator(val *iterationManipulator) {
	for _, item := range iterationManipulators {
		if item.Code == val.Code {
			val.Code = random.RandSeq(80)
			ensureUniqueCodeForIterationManipulator(val)
			return
		}
	}
}

func updateIterationManipulator(c *gin.Context) {
	code := c.Param("code")

	if database == nil {
		defer saveIterationManipulators()
	}
	var data updateIterationManipulatorData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, v := range iterationManipulators {
		if v.Code == code {
			_, err := applyUpdateToIterationManipulator(data, v)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusAccepted, v.Data)
			return
		}
	}
	c.Status(http.StatusNoContent)
}

func applyUpdateToIterationManipulator(data updateIterationManipulatorData, v *iterationManipulator) (dur time.Duration, err error) {
	// verify valid input
	if data.Duration != nil {
		dur, err = parseISO8601Duration(*data.Duration, time.Second)
		if err != nil {
			return
		}
	}

	// apply changes
	if data.Duration != nil {
		v.Data.Duration = *data.Duration
		v.Manipulator.Reset(dur)
	}
	if data.Value != nil {
		v.Data.Value = *data.Value
	}
	if database != nil {
		err = updateManipulatorInDatabase(v.Code, v.Data.Duration, v.Data.Value)
		if err != nil {
			return 0, err
		}
	}
	return
}

func deleteIterationManipulator(c *gin.Context) {
	code := c.Param("code")

	if database == nil {
		defer saveIterationManipulators()
	} else {
		err := deleteManipulatorFromDatabase(code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	for i, v := range iterationManipulators {
		if v.Code == code {
			v.Manipulator.Stop()
			iterationManipulators = append(iterationManipulators[:i], iterationManipulators[i+1:]...)
			c.Status(http.StatusAccepted)
			return
		}
	}
	c.Status(http.StatusOK)
}

func prepareManipulator(route *gin.RouterGroup) {
	route.GET("", findAllIterationManipulators)
	route.GET("/code/:code", findIterationManipulatorByCode)
	route.POST("", createIterationManipulator)
	route.PATCH("/code/:code", updateIterationManipulator)
	route.DELETE("/code/:code", deleteIterationManipulator)
}

package go_gin_pages

import (
	"errors"
	"net/http"
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

type updateiterationManipulator struct {
	Duration *types.ISO8601Duration `json:"Duration"`
	Value    *int                   `json:"Value"`
}

type iterationManipulator struct {
	Code        string
	Data        manipulateIterationData
	Manipulator *time.Ticker
}

type iterationManipulatorCompatibeWithJSON struct {
	Code string
	Data manipulateIterationData
}

func (obj iterationManipulator) ToStructCompatibleWithJSON() iterationManipulatorCompatibeWithJSON {
	return iterationManipulatorCompatibeWithJSON{
		Code: obj.Code,
		Data: obj.Data,
	}
}

var iterationManipulators []*iterationManipulator

func manipulateIteration(obj *iterationManipulator) error {
	for range obj.Manipulator.C {
		iterationMutex.Lock()
		iteration += obj.Data.Value
		if err := saveIteration(); err != nil {
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

func prepareManipulator(route *gin.RouterGroup) {
	route.GET("", findAllIterationManipulators)
	route.GET("/code/:code", findIterationManipulatorByCode)
	route.POST("", createIterationManipulator)
	route.PUT("/code/:code", updateIterationManipulator)
	route.DELETE("/code/:code", deleteIterationManipulator)
}

func findAllIterationManipulators(c *gin.Context) {
	var result []iterationManipulatorCompatibeWithJSON = make([]iterationManipulatorCompatibeWithJSON, 0)
	for _, v := range iterationManipulators {
		result = append(result, v.ToStructCompatibleWithJSON())
	}
	c.JSON(
		http.StatusOK,
		result,
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
	c.JSON(
		http.StatusNoContent,
		gin.H{},
	)
}

func createIterationManipulator(c *gin.Context) {
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
	iterationManipulators = append(iterationManipulators, &iterationManipulator)
	go manipulateIteration(&iterationManipulator)
	c.JSON(
		http.StatusCreated,
		iterationManipulator.ToStructCompatibleWithJSON(),
	)
}

func updateIterationManipulator(c *gin.Context) {
	var data updateiterationManipulator
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	code := c.Param("code")
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
	c.JSON(
		http.StatusNoContent,
		gin.H{},
	)
}

func applyUpdateToIterationManipulator(data updateiterationManipulator, v *iterationManipulator) (dur time.Duration, err error) {
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
	return
}

func deleteIterationManipulator(c *gin.Context) {
	code := c.Param("code")
	for i, v := range iterationManipulators {
		if v.Code == code {
			v.Manipulator.Stop()
			iterationManipulators = append(iterationManipulators[:i], iterationManipulators[i+1:]...)
			c.JSON(
				http.StatusAccepted,
				gin.H{},
			)
			return
		}
	}
	c.JSON(
		http.StatusOK,
		gin.H{},
	)
}

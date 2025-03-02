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

func FindAllIterationManipulator(c *gin.Context) {
	var result []iterationManipulatorCompatibeWithJSON = make([]iterationManipulatorCompatibeWithJSON, 0)
	for _, v := range iterationManipulators {
		result = append(result, v.ToStructCompatibleWithJSON())
	}
	c.JSON(
		http.StatusOK,
		result,
	)
}

func FindIterationManipulatorByCode(c *gin.Context) {
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

func CreateIterationManipulator(c *gin.Context) {
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

func UpdateIterationManipulator(c *gin.Context) {
	var data updateiterationManipulator
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var dur time.Duration
	var err error

	code := c.Param("code")
	for _, v := range iterationManipulators {
		if v.Code == code {
			// verify valid input
			if data.Duration != nil {
				dur, err = parseISO8601Duration(*data.Duration, time.Second)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			c.JSON(
				http.StatusAccepted,
				gin.H{},
			)
			return
		}
	}
	c.JSON(
		http.StatusNoContent,
		gin.H{},
	)
}

func DeleteIterationManipulator(c *gin.Context) {
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

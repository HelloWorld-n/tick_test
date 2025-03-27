package go_gin_pages

import (
	"net/http"
	"time"

	"tick_test/repository"
	"tick_test/types"
	"tick_test/utils/random"

	"github.com/gin-gonic/gin"
)

func findAllIterationManipulators(c *gin.Context) {
	c.JSON(
		http.StatusOK,
		repository.IterationManipulators,
	)
}

func findIterationManipulatorByCode(c *gin.Context) {
	code := c.Param("code")

	for _, v := range repository.IterationManipulators {
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
	var data repository.ManipulateIterationData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	dur, err := types.ParseISO8601Duration(data.Duration, time.Second)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	ticker := time.NewTicker(dur)
	defer func() {
		if err != nil {
			ticker.Stop()
		}
	}()

	iterationManipulator := repository.IterationManipulator{
		Code:        random.RandSeq(80),
		Data:        data,
		Manipulator: ticker,
	}

	repository.IterationManipulatorMutex.Lock()
	// Ensure code is unique
	for {
		unique := true
		for _, item := range repository.IterationManipulators {
			if item.Code == iterationManipulator.Code {
				unique = false
				break
			}
		}
		if unique {
			break
		}
		iterationManipulator.Code = random.RandSeq(80)
	}
	repository.IterationManipulatorMutex.Unlock()

	err = repository.SaveIterationManipulatorToDatabase(&iterationManipulator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	repository.IterationManipulatorMutex.Lock()
	repository.IterationManipulators = append(repository.IterationManipulators, &iterationManipulator)
	repository.IterationManipulatorMutex.Unlock()

	go repository.ManipulateIteration(&iterationManipulator)
	c.JSON(
		http.StatusCreated,
		iterationManipulator,
	)
}

func updateIterationManipulator(c *gin.Context) {
	code := c.Param("code")

	var data repository.UpdateIterationManipulatorData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	for _, v := range repository.IterationManipulators {
		if v.Code == code {
			_, err := repository.ApplyUpdateToIterationManipulator(data, v)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
				return
			}
			c.JSON(http.StatusAccepted, v.Data)
			return
		}
	}
	c.Status(http.StatusNoContent)
}

func deleteIterationManipulator(c *gin.Context) {
	code := c.Param("code")

	if !repository.IsDatabaseEnabled() {
		defer repository.SaveIterationManipulators()
	} else {
		err := repository.DeleteManipulatorFromDatabase(code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
	}

	repository.IterationManipulatorMutex.Lock()
	defer repository.IterationManipulatorMutex.Unlock()
	for i, v := range repository.IterationManipulators {
		if v.Code == code {
			v.Manipulator.Stop()
			repository.IterationManipulators = append(repository.IterationManipulators[:i], repository.IterationManipulators[i+1:]...)
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

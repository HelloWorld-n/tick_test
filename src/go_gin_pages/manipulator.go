package go_gin_pages

import (
	"net/http"
	"time"

	"tick_test/repository"
	"tick_test/types"
	"tick_test/utils/random"

	"github.com/gin-gonic/gin"
)

type manipulatorHandler struct {
	repo repository.ManipulatorRepository
}

func NewManipulatorHandler(manipulatorRepo repository.ManipulatorRepository) (res *manipulatorHandler) {
	return &manipulatorHandler{
		repo: manipulatorRepo,
	}
}

func (mh *manipulatorHandler) findAllIterationManipulatorsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, repository.IterationManipulators)
	}
}

func (mh *manipulatorHandler) findIterationManipulatorByCodeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")
		for _, v := range repository.IterationManipulators {
			if v.Code == code {
				c.JSON(http.StatusOK, v.Data)
				return
			}
		}
		c.Status(http.StatusNoContent)
	}
}

func (mh *manipulatorHandler) createIterationManipulatorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
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

		err = mh.repo.SaveIterationManipulatorToDatabase(&iterationManipulator)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		repository.IterationManipulatorMutex.Lock()
		repository.IterationManipulators = append(repository.IterationManipulators, &iterationManipulator)
		repository.IterationManipulatorMutex.Unlock()

		go repository.ManipulateIteration(&iterationManipulator)
		c.JSON(http.StatusCreated, iterationManipulator)
	}
}

func (mh *manipulatorHandler) updateIterationManipulatorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")
		var data repository.UpdateIterationManipulatorData
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		for _, v := range repository.IterationManipulators {
			if v.Code == code {
				_, err := mh.repo.ApplyUpdateToIterationManipulator(data, v)
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
}

func (mh *manipulatorHandler) deleteIterationManipulatorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Param("code")

		if !mh.repo.IsDatabaseEnabled() {
			defer mh.repo.SaveIterationManipulators()
		} else {
			err := mh.repo.DeleteManipulatorFromDatabase(code)
			if err != nil {
        returnError(c, err)
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
}

func (mh *manipulatorHandler) prepareManipulator(route *gin.RouterGroup) {
	route.GET("", mh.findAllIterationManipulatorsHandler())
	route.GET("/code/:code", mh.findIterationManipulatorByCodeHandler())
	route.POST("", mh.createIterationManipulatorHandler())
	route.PATCH("/code/:code", mh.updateIterationManipulatorHandler())
	route.DELETE("/code/:code", mh.deleteIterationManipulatorHandler())
}

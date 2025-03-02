package main

import (
	"tick_test/go_gin_pages"

	"github.com/gin-gonic/gin"
)

func main() {
	go_gin_pages.Prepare()

	ginServer := gin.Default()
	ginServer.GET("/", go_gin_pages.Index)
	ginServer.GET("/manipulator", go_gin_pages.FindAllIterationManipulator)
	ginServer.GET("/manipulator/code/:code", go_gin_pages.FindIterationManipulatorByCode)
	ginServer.POST("/manipulator", go_gin_pages.CreateIterationManipulator)
	ginServer.PUT("/manipulator/code/:code", go_gin_pages.UpdateIterationManipulator)
	ginServer.DELETE("/manipulator/code/:code", go_gin_pages.DeleteIterationManipulator)
	ginServer.Run("127.0.0.1:4041")
}

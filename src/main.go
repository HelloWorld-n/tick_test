package main

import (
	"tick_test/go_gin_pages"

	"github.com/gin-gonic/gin"
)

func main() {
	ginServer := gin.Default()
	ginServer.UseRawPath = true
	go_gin_pages.Prepare(ginServer)
	ginServer.Run("127.0.0.1:4041")
}

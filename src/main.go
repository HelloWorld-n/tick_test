package main

import (
	"tick_test/go_gin_pages"

	"github.com/gin-gonic/gin"
)

func main() {
	url := "127.0.0.1:4041"
	ginServer := gin.Default()
	ginServer.UseRawPath = true
	go_gin_pages.Prepare(ginServer, url)
	ginServer.Run(url)
}

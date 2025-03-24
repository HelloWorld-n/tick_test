package main

import (
	"tick_test/go_gin_pages"

	"github.com/gin-gonic/gin"
)

func main() {
	url, _ := go_gin_pages.DetermineURL()
	ginServer := gin.Default()
	ginServer.UseRawPath = true
	go_gin_pages.Prepare(ginServer, url)
	ginServer.Run(url)
}

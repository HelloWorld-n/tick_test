package main

import (
	"os"
	"tick_test/go_gin_pages"

	"github.com/gin-gonic/gin"
)

func main() {
	url, err := go_gin_pages.UseConfigToDetermineURL()
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	ginServer := gin.Default()
	ginServer.UseRawPath = true
	go_gin_pages.Prepare(ginServer, url)
	ginServer.Run(url)
}

package main

import (
	"flag"
	"fmt"
	"os"
	"tick_test/go_gin_pages"
	"tick_test/internal/config"

	"github.com/gin-gonic/gin"
)

func main() {
	configPath := flag.String("c", "config.yaml", "Path to config file")
	flag.Parse()
	fmt.Println("PATH", *configPath)
	cfg, err := config.GetConfig(*configPath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
	url := go_gin_pages.UseConfigToDetermineURL(cfg)
	ginServer := gin.Default()
	ginServer.UseRawPath = true
	go_gin_pages.Prepare(ginServer, url)
	ginServer.Run(url)
}

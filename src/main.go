package main

import (
	"flag"
	"fmt"
	"os"
	"tick_test/go_gin_pages"
	"tick_test/internal/config"
	"tick_test/repository"
	"tick_test/utils/jwt"

	"github.com/gin-gonic/gin"
)

func setupRepository() (repo repository.Repository, err error) {
	dbPath, err := repository.LoadDatabasePath()
	if err != nil {
		return nil, fmt.Errorf("could not load database path: %w", err)
	}

	db, err := repository.NewDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %w", err)
	}

	return repository.NewRepo(db), nil
}

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

	jwt.SetSecretKey([]byte(os.Getenv("JWT_SECRET_KEY")))
	repo, err := setupRepository()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Repository setup failed:", err)
		os.Exit(1)
	}

	ginServer := gin.Default()
	ginServer.UseRawPath = true

	go_gin_pages.Prepare(ginServer, url, repo)
	ginServer.Run(url)
}

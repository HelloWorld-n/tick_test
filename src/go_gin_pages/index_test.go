package go_gin_pages

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"tick_test/internal/config"
)

var url string
var client *http.Client
var adminCredentialsPath = "../../.config/adminCredentials.json"

var adminCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func setupIndex() {
	configPath := flag.String("c", "config.yaml", "Path to config file")
	cfg, err := config.GetConfig(*configPath)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	url = UseConfigToDetermineURL(cfg)
	client = &http.Client{}
	adminCredentialsFile, err := os.Open(adminCredentialsPath)
	if err != nil {
		panic(fmt.Errorf("can not open file %v for reason %w", adminCredentialsFile, err))
	}
	adminCredentialsData, err := io.ReadAll(adminCredentialsFile)
	if err != nil {
		panic(fmt.Errorf("can not open file %v for reason %w", adminCredentialsFile, err))
	}
	json.Unmarshal(adminCredentialsData, &adminCredentials)
	defer adminCredentialsFile.Close()
}

func setup() {
	setupAccount()
	setupBook()
}

func fmtPrintlnRespone(resp *http.Response) {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	resp.Body = io.NopCloser(bytes.NewReader(data))
	fmt.Println()
	fmt.Println(resp.Status)
	fmt.Println(string(data))
	fmt.Println()
}

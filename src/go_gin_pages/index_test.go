package go_gin_pages_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"tick_test/go_gin_pages"
)

var url string
var client *http.Client
var adminCredentialsPath = "../../.config/adminCredentials.json"

var adminCredentials struct {
	Username string `json:"Username"`
	Password string `json:"Password"`
}

func setupIndex() {
	url, _ = go_gin_pages.DetermineURL()
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

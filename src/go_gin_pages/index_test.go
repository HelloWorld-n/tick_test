package go_gin_pages_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"
	"tick_test/go_gin_pages"
)

var url string
var client *http.Client

func setupIndex() {
	url, _ = go_gin_pages.DetermineURL()
	client = &http.Client{}
}

func setup() {
	setupAccount()
	setupBook()
}

func TestAll(t *testing.T) {
	TestAllAccounts(t)
	TestAllBooks(t)
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

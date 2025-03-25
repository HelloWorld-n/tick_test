package go_gin_pages_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"tick_test/go_gin_pages"
	"tick_test/utils/random"

	"gopkg.in/go-playground/assert.v1"
)

var url string
var client *http.Client

type accountData struct {
	Username  string
	Password  string
	Role      string
	IsCreated bool
}

var accounts []accountData

func setup() {
	url, _ = go_gin_pages.DetermineURL()
	client = &http.Client{}
	accounts = make([]accountData, 0)
}

func createAccounts(n int) (prevN int) {
	prevN = len(accounts)
	for range n {
		account := accountData{
			Username: "IDENTIFICATION/" + random.RandSeq(15),
			Password: "VERIFICATION/" + random.RandSeq(30),
			Role:     "User",
		}
		accounts = append(accounts, account)
	}
	return
}

func fmtPrintlnRespone(resp *http.Response) {
	var data = make([]byte, 65536)
	n, _ := resp.Body.Read(data)
	fmt.Println()
	fmt.Println(resp.Status)
	fmt.Println(string(data[:n]))
	fmt.Println()
}

func accountCreator(accPos int, isSamePassword bool) (result func(t *testing.T)) {
	return func(t *testing.T) {
		samePassword := accounts[accPos].Password
		if !isSamePassword {
			samePassword = "VERIFICATION/" + random.RandSeq(30)
		}
		body, _ := json.Marshal(
			go_gin_pages.AccountPostData{
				Username:     accounts[accPos].Username,
				Password:     accounts[accPos].Password,
				SamePassword: samePassword,
				Role:         accounts[accPos].Role,
			},
		)
		req, err := http.NewRequest(http.MethodPost, "http://"+url+"/account/register", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		fmtPrintlnRespone(resp)

		if !isSamePassword {
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
			return
		}

		if accounts[accPos].IsCreated {
			assert.Equal(t, http.StatusConflict, resp.StatusCode)
			return
		}

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		accounts[accPos].IsCreated = true
	}
}

func accountDeleter(accPos int) (result func(t *testing.T)) {
	return func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, "http://"+url+"/account/delete", bytes.NewBuffer([]byte{}))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Username", accounts[accPos].Username)
		req.Header.Set("Password", accounts[accPos].Password)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		fmtPrintlnRespone(resp)

		if accounts[accPos].IsCreated {
			assert.Equal(t, http.StatusAccepted, resp.StatusCode)
			accounts[accPos].IsCreated = false
			return
		}

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}

func TestCreateAccount(t *testing.T) {
	setup()
	accPos := createAccounts(2)

	fmt.Println("🟪 Success")
	t.Run("Success", accountCreator(accPos, true))

	fmt.Println("🟪 Failure/ExistingUser")
	t.Run("Failure/ExistingUser", accountCreator(accPos, true))

	fmt.Println("🟪 Failure/DifferentPasswords")
	t.Run("Failure/DifferentPasswords", accountCreator(accPos+1, false))

	fmt.Println("🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫")
}

func TestLogin(t *testing.T) {
	// TODO: implement
}

func TestPatchAccount(t *testing.T) {
	// TODO: implement
}

func TestDeleteAccount(t *testing.T) {
	setup()
	accPos := createAccounts(2)

	fmt.Println("🟫 -/Creating")

	for i := range 2 {
		t.Run("-", accountCreator(accPos+i, true))
	}

	fmt.Println("🟪 Success/ExistingUser")

	for i := range 2 {
		t.Run("-", accountDeleter(accPos+i))
	}

	fmt.Println("🟪 Failure/NonExistingUser")
	for i := range 2 {
		t.Run("-", accountDeleter(accPos+i))
	}

	fmt.Println("🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫🟫")
}

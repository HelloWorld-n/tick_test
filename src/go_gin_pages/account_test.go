package go_gin_pages_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	typeDefs "tick_test/types"
	"tick_test/utils/random"

	"gopkg.in/go-playground/assert.v1"
)

type accountData struct {
	Username  string
	Password  string
	Role      string
	IsCreated bool
}

var accounts []accountData

func setupAccount() {
	setupIndex()
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

func accountCreator(accPos int, isSamePassword bool) (result func(t *testing.T)) {
	return func(t *testing.T) {
		samePassword := accounts[accPos].Password
		if !isSamePassword {
			samePassword = "VERIFICATION/" + random.RandSeq(30)
		}
		body, _ := json.Marshal(
			typeDefs.AccountPostData{
				Username:     accounts[accPos].Username,
				Password:     accounts[accPos].Password,
				SamePassword: samePassword,
				Role:         accounts[accPos].Role,
			},
		)
		req, err := http.NewRequest(http.MethodPost, "http://"+url+"/account/register", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
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

func accountPatcher(accPos int, newPassword string, confirmNewPassword string) (result func(t *testing.T)) {
	return func(t *testing.T) {
		token, err := accountLogin(accPos, true)
		if err != nil {
			t.Fatalf("failed to login: %v", err)
		}
		patchPayload := typeDefs.AccountPatchData{
			Password:     newPassword,
			SamePassword: confirmNewPassword,
		}
		body, _ := json.Marshal(patchPayload)
		req, err := http.NewRequest(http.MethodPatch, "http://"+url+"/account/modify", bytes.NewBuffer(body))
		if err != nil {
			t.Fatalf("failed to create patch request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Token", token)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("patch request failed: %v", err)
		}
		defer resp.Body.Close()

		fmtPrintlnRespone(resp)
		if newPassword == confirmNewPassword {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		} else {
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		}
	}
}

func accountLogin(accPos int, isCorrectPassword bool) (token string, err error) {
	body := []byte(`{}`)
	req, err := http.NewRequest(http.MethodPost, "http://"+url+"/account/login", bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create login request: %v", err)
	}
	req.Header.Set("Username", accounts[accPos].Username)
	if isCorrectPassword {
		req.Header.Set("Password", accounts[accPos].Password)
	} else {
		req.Header.Set("Password", "FAILED-ATTEMPT")
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("login request failed: %v", err)
	}
	defer resp.Body.Close()

	fmtPrintlnRespone(resp)

	tokenBytes, err := io.ReadAll(resp.Body)
	if err := json.Unmarshal(tokenBytes, &token); err != nil && resp.StatusCode == http.StatusOK {
		return "", fmt.Errorf("error unmarshalling token: %v", err)
	}

	return token, nil
}

func accountLoginer(accPos int, isCorrectPassword bool) func(t *testing.T) {
	return func(t *testing.T) {
		var err error
		_, err = accountLogin(accPos, isCorrectPassword)
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}
	}
}

func accountDeleter(accPos int) (result func(t *testing.T)) {
	return func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, "http://"+url+"/account/delete", bytes.NewBuffer([]byte{}))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
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

	t.Run("Success", accountCreator(accPos, true))

	t.Run("Failure/ExistingUser", accountCreator(accPos, true))

	t.Run("Failure/DifferentPasswords", accountCreator(accPos+1, false))

}

func TestLogin(t *testing.T) {
	setup()
	accPos := createAccounts(1)
	t.Run("Register", accountCreator(accPos, true))

	t.Run("Success", accountLoginer(accPos, true))

	t.Run("Failure/InvalidCredentials", accountLoginer(accPos, false))
}

func TestPatchAccount(t *testing.T) {
	setup()
	accPos := createAccounts(1)
	t.Run("Register", accountCreator(accPos, true))

	newPassword := "VERIFICATION/" + random.RandSeq(30)
	t.Run("Valid Patch", accountPatcher(accPos, newPassword, newPassword))

	accounts[accPos].Password = newPassword
	t.Run("Login With Updated Password", accountLoginer(accPos, true))

	t.Run("Failure/MismatchedPasswords", accountPatcher(accPos, newPassword, "MismatchedPassword"))
}

func TestDeleteAccount(t *testing.T) {
	setup()
	accPos := createAccounts(2)

	for i := range 2 {
		t.Run("CreateForDeletion", accountCreator(accPos+i, true))
	}

	for i := range 2 {
		t.Run("DeleteExisting", accountDeleter(accPos+i))
	}

	for i := range 2 {
		t.Run("DeleteMissing", accountDeleter(accPos+i))
	}

}

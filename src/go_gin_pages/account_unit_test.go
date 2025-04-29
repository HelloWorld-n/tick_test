package go_gin_pages_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	ginPages "tick_test/go_gin_pages"
	"tick_test/go_gin_pages/mocks"
	"tick_test/repository"
	"tick_test/types"
	"tick_test/utils/errDefs"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetAllAccounts(t *testing.T) {
	testCases := []struct {
		name            string
		repo            repository.AccountRepository
		expectedPayload string
		expectedStatus  int
	}{
		{
			name: "Success",
			repo: &mocks.AccountRepositoryMock{
				FindAllAccountsFn: func() ([]types.AccountGetData, error) {
					return []types.AccountGetData{
						{
							Username: "Strale",
							Role:     "admin",
						},
					}, nil
				},
			},
			expectedPayload: `[{"username":"Strale","role":"admin"}]`,
			expectedStatus:  http.StatusOK,
		},
		{
			name: "Fail",
			repo: &mocks.AccountRepositoryMock{
				FindAllAccountsFn: func() ([]types.AccountGetData, error) {
					return nil, errors.New("something happened")
				},
			},
			expectedPayload: `{"Error":"something happened"}`,
			expectedStatus:  http.StatusInternalServerError,
		},
		{
			name: "Empty result set",
			repo: &mocks.AccountRepositoryMock{
				FindAllAccountsFn: func() ([]types.AccountGetData, error) {
					return []types.AccountGetData{}, nil
				},
			},
			expectedPayload: `[]`,
			expectedStatus:  http.StatusOK,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ah := ginPages.NewAccountHandler(tc.repo)
			handler := ah.GetAllAccountsHandler()

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			handler(ctx)

			assert.Equal(t, tc.expectedStatus, w.Code)
			assert.Equal(t, tc.expectedPayload, w.Body.String())
		})
	}
}

func TestLoginHandler_Success(t *testing.T) {
	repo := &mocks.AccountRepositoryMock{
		ConfirmAccountFn: func(username, password string) error {
			if username == "valid" && password == "pass" {
				return nil
			}
			return errors.New("invalid")
		},
		FindUserRoleFn: func(username string) (string, error) {
			return "User", nil
		},
	}

	handler := ginPages.NewAccountHandler(repo).LoginHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", nil)
	c.Request.Header.Set("Username", "valid")
	c.Request.Header.Set("Password", "pass")

	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPostAccountHandler(t *testing.T) {
	testCases := []struct {
		name            string
		repo            *mocks.AccountRepositoryMock
		inputPayload    string
		expectedStatus  int
		expectedPayload string
	}{
		{
			name: "Success",
			repo: &mocks.AccountRepositoryMock{
				SaveAccountFn: func(data *types.AccountPostData) error {
					return nil
				},
			},
			inputPayload:    `{"username":"TESTER","password":"WORKING_PASSWORD","samePassword":"WORKING_PASSWORD","role":"Admin"}`,
			expectedStatus:  http.StatusCreated,
			expectedPayload: `{"username":"TESTER","password":"WORKING_PASSWORD","samePassword":"WORKING_PASSWORD","role":"Admin"}`,
		},
		{
			name: "Success with default role",
			repo: &mocks.AccountRepositoryMock{
				SaveAccountFn: func(data *types.AccountPostData) error {
					return nil
				},
			},
			inputPayload:    `{"username":"TESTER","password":"WORKING_PASSWORD","samePassword":"WORKING_PASSWORD"}`,
			expectedStatus:  http.StatusCreated,
			expectedPayload: `{"username":"TESTER","password":"WORKING_PASSWORD","samePassword":"WORKING_PASSWORD","role":"User"}`,
		},
		{
			name: "Fail - bind error",
			repo: &mocks.AccountRepositoryMock{
				SaveAccountFn: func(data *types.AccountPostData) error {
					return nil
				},
			},
			inputPayload:    `invalid json`,
			expectedStatus:  http.StatusBadRequest,
			expectedPayload: `{"Error":"json: cannot unmarshal string into Go value of type types.AccountPostData"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ah := ginPages.NewAccountHandler(tc.repo)
			handler := ah.PostAccountHandler()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(tc.inputPayload))
			c.Request.Header.Set("Content-Type", "application/json")

			handler(c)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if w.Code != http.StatusBadRequest {
				assert.JSONEq(t, tc.expectedPayload, w.Body.String())
			}
		})
	}
}

func TestPatchAccountHandler(t *testing.T) {
	testCases := []struct {
		name             string
		repo             *mocks.AccountRepositoryMock
		inputPayload     string
		usernameHeader   string
		passwordHeader   string
		expectedStatus   int
		ConfirmAccountFn func(string, string) error
	}{
		{
			name: "Success",
			repo: &mocks.AccountRepositoryMock{
				ConfirmAccountFn: func(string, string) error {
					return nil
				},
				UpdateExistingAccountFn: func(string, *types.AccountPatchData) error {
					return nil
				},
			},
			inputPayload:   `{"password":"NEW_PASSWORD","samePassword":"NEW_PASSWORD","email":"test@example.com"}`,
			usernameHeader: "TESTER",
			passwordHeader: "ORIGINAL_PASSWORD",
			expectedStatus: http.StatusOK,
		},
		{
			name: "Fail - ConfirmUser error",
			repo: &mocks.AccountRepositoryMock{
				ConfirmAccountFn: func(string, string) error {
					return errDefs.ErrUnauthorized
				},
			},
			inputPayload:   `{"password":"NEW_PASSWORD","samePassword":"NEW_PASSWORD","email":"test@example.com"}`,
			usernameHeader: "TESTER",
			passwordHeader: "ORIGINAL_PASSWORD",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Fail - BindJSON error",
			repo: &mocks.AccountRepositoryMock{
				ConfirmAccountFn: func(string, string) error {
					return nil
				},
			},
			inputPayload:   `invalid json`,
			usernameHeader: "TESTER",
			passwordHeader: "ORIGINAL_PASSWORD",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ah := ginPages.NewAccountHandler(tc.repo)
			handler := ah.PatchAccountHandler()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/modify", bytes.NewBufferString(tc.inputPayload))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Username", tc.usernameHeader)
			c.Request.Header.Set("Password", tc.passwordHeader)

			handler(c)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestPatchPromoteAccountHandler(t *testing.T) {
	testCases := []struct {
		name           string
		repo           *mocks.AccountRepositoryMock
		inputPayload   string
		expectedStatus int
		userRole       string
		confirmUserErr error
	}{
		{
			name: "Success",
			repo: &mocks.AccountRepositoryMock{
				ConfirmAccountFn: func(string, string) error { return nil },
				FindUserRoleFn:   func(username string) (string, error) { return "Admin", nil },
				PromoteExistingAccountFn: func(data *types.AccountPatchPromoteData) error {
					return nil
				},
			},
			inputPayload:   `{"username":"PROMOTEÉ","role":"Admin"}`,
			expectedStatus: http.StatusOK,
			userRole:       "Admin",
			confirmUserErr: nil,
		},
		{
			name: "Fail - Not Admin",
			repo: &mocks.AccountRepositoryMock{
				ConfirmAccountFn: func(string, string) error { return nil },
				FindUserRoleFn:   func(username string) (string, error) { return "User", nil },
				PromoteExistingAccountFn: func(data *types.AccountPatchPromoteData) error {
					return nil
				},
			},
			inputPayload:   `{"username":"PROMOTEÉ","role":"Admin"}`,
			expectedStatus: http.StatusUnauthorized,
			userRole:       "User",
			confirmUserErr: nil,
		},
		{
			name: "Fail - BindJSON error",
			repo: &mocks.AccountRepositoryMock{
				ConfirmAccountFn: func(string, string) error { return nil },
				FindUserRoleFn:   func(username string) (string, error) { return "Admin", nil },
			},
			inputPayload:   `invalid json`,
			expectedStatus: http.StatusBadRequest,
			userRole:       "Admin",
			confirmUserErr: nil,
		},
		{
			name: "Fail - Confirm Account Error",
			repo: &mocks.AccountRepositoryMock{
				ConfirmAccountFn: func(string, string) error { return errDefs.ErrUnauthorized },
			},
			inputPayload:   `{"username":"PROMOTEÉ","role":"Admin"}`,
			expectedStatus: http.StatusUnauthorized,
			userRole:       "Admin",
			confirmUserErr: errDefs.ErrUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ah := ginPages.NewAccountHandler(tc.repo)
			handler := ah.PatchPromoteAccountHandler()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPatch, "/promote", bytes.NewBufferString(tc.inputPayload))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Request.Header.Set("Username", "SUDOER")
			c.Request.Header.Set("Password", "VERIFICATION")

			handler(c)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

func TestDeleteAccountHandler(t *testing.T) {
	testCases := []struct {
		name             string
		repo             *mocks.AccountRepositoryMock
		usernameHeader   string
		passwordHeader   string
		expectedStatus   int
		userExists       bool
		confirmUserErr   error
		deleteAccountErr error
	}{
		{
			name: "Success",
			repo: &mocks.AccountRepositoryMock{
				UserExistsFn: func(username string) (bool, error) { return true, nil },
				ConfirmAccountFn: func(string, string) error {
					return nil
				},
				DeleteAccountFn: func(username string) error { return nil },
			},
			usernameHeader:   "TESTER",
			passwordHeader:   "VERIFICATION",
			expectedStatus:   http.StatusAccepted,
			userExists:       true,
			confirmUserErr:   nil,
			deleteAccountErr: nil,
		},
		{
			name: "Success - User Does Not Exist",
			repo: &mocks.AccountRepositoryMock{
				UserExistsFn: func(username string) (bool, error) { return false, nil },
			},
			usernameHeader:   "TESTER",
			passwordHeader:   "VERIFICATION",
			expectedStatus:   http.StatusOK,
			userExists:       false,
			confirmUserErr:   nil,
			deleteAccountErr: nil,
		},
		{
			name: "Fail - ConfirmUser Error",
			repo: &mocks.AccountRepositoryMock{
				UserExistsFn: func(username string) (bool, error) { return true, nil },
				ConfirmAccountFn: func(string, string) error {
					return errDefs.ErrUnauthorized
				},
			},
			usernameHeader:   "TESTER",
			expectedStatus:   http.StatusUnauthorized,
			userExists:       true,
			confirmUserErr:   errDefs.ErrUnauthorized,
			deleteAccountErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ah := ginPages.NewAccountHandler(tc.repo)
			handler := ah.DeleteAccountHandler()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/delete", nil)
			c.Request.Header.Set("Username", tc.usernameHeader)
			c.Request.Header.Set("Password", tc.passwordHeader)

			handler(c)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}

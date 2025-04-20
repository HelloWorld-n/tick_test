package go_gin_pages_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	ginPages "tick_test/go_gin_pages"
	"tick_test/go_gin_pages/mocks"
	"tick_test/repository"
	"tick_test/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)


func TestGetAllAccounts(mainT *testing.T) {
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
		mainT.Run(tc.name, func(t *testing.T) {
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
	}

	handler := ginPages.NewAccountHandler(repo).LoginHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", nil)
	c.Request.Header.Set("Username", "valid")
	c.Request.Header.Set("Password", "pass")

	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Len(t, w.Body.String(), 82) // token length
}

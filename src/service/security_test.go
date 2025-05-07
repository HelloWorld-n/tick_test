package service_test

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"tick_test/go_gin_pages/mocks"
	"tick_test/service"
	"tick_test/types"
	"tick_test/utils/errDefs"
	"tick_test/utils/jwt"
)

var validateTokenFn func(token string) (*jwt.Claims, error)

func TestCheckRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwt.SetSecretKey([]byte("RU5DT0RFRF9TRUNSRVRfVEVYVA=="))
	validToken, _ := jwt.GenerateToken("ADMINISTRATOR", types.AdminRole, 30*time.Minute)
	invalidToken := "eyJhbGciOk5vbmV9.e30-.e30-"

	tests := []struct {
		name                      string
		token                     string
		mockClaims                *jwt.Claims
		mockTokenErr              error
		mockAccountId             int
		mockRepoErr               error
		expectedSuccess           bool
		expectedErr               error
		findAccountIdByUsernameFn func(string) (int64, error)
	}{
		{
			name:  "valid token and matching role",
			token: validToken,
			mockClaims: &jwt.Claims{
				Username: "ADMINISTRATOR",
				Role:     types.AdminRole,
			},
			mockAccountId:   1,
			expectedSuccess: true,
			expectedErr:     nil,
			findAccountIdByUsernameFn: func(string) (int64, error) {
				return 1, nil
			},
		},
		{
			name:            "invalid token",
			token:           invalidToken,
			mockTokenErr:    errors.New("invalid"),
			expectedSuccess: false,
			expectedErr:     errDefs.ErrUnauthorized,
		},
		{
			name:  "repo error",
			token: validToken,
			mockClaims: &jwt.Claims{
				Username: "SUDOER",
				Role:     types.AdminRole,
			},
			mockRepoErr:     errors.New("db fail"),
			expectedSuccess: false,
			expectedErr:     errDefs.ErrUnauthorized,
			findAccountIdByUsernameFn: func(string) (int64, error) {
				return 0, errors.New("db fail")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.AccountRepositoryMock{
				FindAccountIdByUsernameFn: tt.findAccountIdByUsernameFn,
			}

			svc := service.SecurityService{AccRepo: mockRepo}

			validateTokenFn = func(_ string) (*jwt.Claims, error) {
				if tt.mockTokenErr != nil {
					return nil, tt.mockTokenErr
				}
				return tt.mockClaims, nil
			}

			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("User-Token", tt.token)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			acc, ok, err := svc.CheckRole(ctx, types.AdminRole)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.False(t, ok)
			} else {
				assert.NoError(t, err)
				assert.True(t, ok)
				assert.Equal(t, tt.mockClaims.Username, acc.Username)
				assert.Equal(t, types.AdminRole, acc.Role)
			}
		})
	}
}

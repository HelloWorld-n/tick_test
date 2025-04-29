package jwt_test

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	jwtpkg "tick_test/utils/jwt"
)

const (
	testUsername  = "testuser"
	testAdminUser = "admin"
	testRole      = "User"
	testAdminRole = "Admin"
	testSecret    = "test-secret-key-for-jwt-unit-tests"
)

func TestGenerateToken_UnsetSecretKey(t *testing.T) {
	originalKey := jwtpkg.GetSecretKey()
	defer jwtpkg.SetSecretKey(originalKey)

	jwtpkg.SetSecretKey(nil)

	token, err := jwtpkg.GenerateToken(testUsername, testRole, time.Hour)
	if err == nil {
		t.Fatalf("Expected error when secret key is unset, got no error")
	}
	if token != "" {
		t.Errorf("Expected empty token when secret key is unset, got %q", token)
	}
}

func TestValidateToken_UnsetSecretKey(t *testing.T) {
	originalKey := jwtpkg.GetSecretKey()
	defer jwtpkg.SetSecretKey(originalKey)

	jwtpkg.SetSecretKey([]byte(testSecret))
	token, err := jwtpkg.GenerateToken(testUsername, testRole, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v, expected no error", err)
	}

	jwtpkg.SetSecretKey(nil)
	_, err = jwtpkg.ValidateToken(token)
	if err == nil {
		t.Fatalf("Expected error when secret key is unset during token validation, got no error")
	}
}

func createTestToken(t *testing.T, username, role string, expireIn time.Duration) string {
	t.Helper()
	claims := jwt.MapClaims{
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(expireIn).Unix(),
		"iat":      time.Now().Unix(),
		"issuer":   "test",
		"subject":  username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	return signedToken
}

func createModifiedToken(t *testing.T, username, role string, modifySignature bool) string {
	t.Helper()
	token := createTestToken(t, username, role, time.Hour)

	if modifySignature {
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			t.Fatalf("Token does not have expected format")
		}

		sig := parts[2]
		if len(sig) > 0 {
			modChar := byte(sig[0] + 1)
			parts[2] = string(modChar) + sig[1:]
		}

		return strings.Join(parts, ".")
	}

	return token
}

func TestGenerateToken_Success(t *testing.T) {
	originalKey := jwtpkg.GetSecretKey()
	defer jwtpkg.SetSecretKey(originalKey)

	jwtpkg.SetSecretKey([]byte(testSecret))

	token, err := jwtpkg.GenerateToken(testUsername, testRole, time.Hour)

	if err != nil {
		t.Fatalf("GenerateToken() error = %v, expected no error", err)
	}

	if token == "" {
		t.Fatalf("GenerateToken() returned empty token")
	}

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})

	if err != nil {
		t.Fatalf("Failed to parse generated token: %v", err)
	}

	if !parsedToken.Valid {
		t.Errorf("Generated token is invalid")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("Failed to extract claims from token")
	}

	username, ok := claims["username"].(string)
	if !ok || username != testUsername {
		t.Errorf("Token username claim = %v, want %v", username, testUsername)
	}

	role, ok := claims["role"].(string)
	if !ok || role != testRole {
		t.Errorf("Token role claim = %v, want %v", role, testRole)
	}
}

func TestGenerateToken_InvalidInputs(t *testing.T) {
	originalKey := jwtpkg.GetSecretKey()
	defer jwtpkg.SetSecretKey(originalKey)

	jwtpkg.SetSecretKey([]byte(testSecret))

	tests := []struct {
		name     string
		username string
		role     string
		expireIn time.Duration
		wantErr  bool
	}{
		{
			name:     "Empty username",
			username: "",
			role:     testRole,
			expireIn: time.Hour,
			wantErr:  true,
		},
		{
			name:     "Empty role",
			username: testUsername,
			role:     "",
			expireIn: time.Hour,
			wantErr:  true,
		},
		{
			name:     "Zero expiration",
			username: testUsername,
			role:     testRole,
			expireIn: 0,
			wantErr:  false,
		},
		{
			name:     "Negative expiration",
			username: testUsername,
			role:     testRole,
			expireIn: -time.Hour,
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token, err := jwtpkg.GenerateToken(tc.username, tc.role, tc.expireIn)

			if (err != nil) != tc.wantErr {
				t.Fatalf("GenerateToken() error = %v, wantErr %v", err, tc.wantErr)
			}

			if !tc.wantErr && token == "" {
				t.Errorf("GenerateToken() returned empty token")
			}
		})
	}
}

func TestValidateToken_Success(t *testing.T) {
	originalKey := jwtpkg.GetSecretKey()
	defer jwtpkg.SetSecretKey(originalKey)

	jwtpkg.SetSecretKey([]byte(testSecret))

	tokenString := createTestToken(t, testUsername, testRole, time.Hour)

	claims, err := jwtpkg.ValidateToken(tokenString)

	if err != nil {
		t.Fatalf("ValidateToken() error = %v, expected no error", err)
	}

	if claims.Username != testUsername {
		t.Errorf("Username = %v, want %v", claims.Username, testUsername)
	}

	if claims.Role != testRole {
		t.Errorf("Role = %v, want %v", claims.Role, testRole)
	}
}

func TestValidateToken_Failures(t *testing.T) {
	originalKey := jwtpkg.GetSecretKey()
	defer jwtpkg.SetSecretKey(originalKey)

	jwtpkg.SetSecretKey([]byte(testSecret))

	tests := []struct {
		name      string
		tokenFunc func(t *testing.T) string
		wantErr   string
	}{
		{
			name: "Expired token",
			tokenFunc: func(t *testing.T) string {
				return createTestToken(t, testUsername, testRole, -time.Second)
			},
			wantErr: "expired",
		},
		{
			name: "Empty token",
			tokenFunc: func(t *testing.T) string {
				return ""
			},
			wantErr: "invalid",
		},
		{
			name: "Invalid format",
			tokenFunc: func(t *testing.T) string {
				return "not.a.valid.token"
			},
			wantErr: "invalid",
		},
		{
			name: "Modified signature",
			tokenFunc: func(t *testing.T) string {
				return createModifiedToken(t, testUsername, testRole, true)
			},
			wantErr: "invalid",
		},
		{
			name: "Token without username",
			tokenFunc: func(t *testing.T) string {
				claims := jwt.MapClaims{
					"role": testRole,
					"exp":  time.Now().Add(time.Hour).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				signedToken, err := token.SignedString([]byte(testSecret))
				if err != nil {
					t.Fatalf("Failed to create token: %v", err)
				}
				return signedToken
			},
			wantErr: "missing",
		},
		{
			name: "Token without role",
			tokenFunc: func(t *testing.T) string {
				claims := jwt.MapClaims{
					"username": testUsername,
					"exp":      time.Now().Add(time.Hour).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				signedToken, err := token.SignedString([]byte(testSecret))
				if err != nil {
					t.Fatalf("Failed to create token: %v", err)
				}
				return signedToken
			},
			wantErr: "missing",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tokenString := tc.tokenFunc(t)

			_, err := jwtpkg.ValidateToken(tokenString)

			if err == nil {
				t.Errorf("ValidateToken() expected error containing '%s', got nil", tc.wantErr)
				return
			}

			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("ValidateToken() error = %v, want it to contain '%s'", err, tc.wantErr)
			}
		})
	}
}

func TestGetUserFromToken(t *testing.T) {
	originalKey := jwtpkg.GetSecretKey()
	defer jwtpkg.SetSecretKey(originalKey)

	jwtpkg.SetSecretKey([]byte(testSecret))

	tests := []struct {
		name      string
		tokenFunc func(t *testing.T) string
		wantUser  string
		wantErr   bool
	}{
		{
			name: "Valid token",
			tokenFunc: func(t *testing.T) string {
				return createTestToken(t, testUsername, testRole, time.Hour)
			},
			wantUser: testUsername,
			wantErr:  false,
		},
		{
			name: "Expired token",
			tokenFunc: func(t *testing.T) string {
				return createTestToken(t, testUsername, testRole, -time.Second)
			},
			wantUser: "",
			wantErr:  true,
		},
		{
			name: "Invalid token",
			tokenFunc: func(t *testing.T) string {
				return "invalid.token"
			},
			wantUser: "",
			wantErr:  true,
		},
		{
			name: "Token without username",
			tokenFunc: func(t *testing.T) string {
				claims := jwt.MapClaims{
					"role": testRole,
					"exp":  time.Now().Add(time.Hour).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				signedToken, err := token.SignedString([]byte(testSecret))
				if err != nil {
					t.Fatalf("Failed to create token: %v", err)
				}
				return signedToken
			},
			wantUser: "",
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tokenString := tc.tokenFunc(t)

			username, err := jwtpkg.GetUserFromToken(tokenString)

			if (err != nil) != tc.wantErr {
				t.Errorf("GetUserFromToken() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if username != tc.wantUser {
				t.Errorf("GetUserFromToken() username = %v, want %v", username, tc.wantUser)
			}
		})
	}
}

func TestGetRoleFromToken(t *testing.T) {
	originalKey := jwtpkg.GetSecretKey()
	defer jwtpkg.SetSecretKey(originalKey)

	jwtpkg.SetSecretKey([]byte(testSecret))

	tests := []struct {
		name      string
		tokenFunc func(t *testing.T) string
		wantRole  string
		wantErr   bool
	}{
		{
			name: "Valid token",
			tokenFunc: func(t *testing.T) string {
				return createTestToken(t, testUsername, testRole, time.Hour)
			},
			wantRole: testRole,
			wantErr:  false,
		},
		{
			name: "Admin token",
			tokenFunc: func(t *testing.T) string {
				return createTestToken(t, testAdminUser, testAdminRole, time.Hour)
			},
			wantRole: testAdminRole,
			wantErr:  false,
		},
		{
			name: "Expired token",
			tokenFunc: func(t *testing.T) string {
				return createTestToken(t, testUsername, testRole, -time.Second)
			},
			wantRole: "",
			wantErr:  true,
		},
		{
			name: "Token without role",
			tokenFunc: func(t *testing.T) string {
				claims := jwt.MapClaims{
					"username": testUsername,
					"exp":      time.Now().Add(time.Hour).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				signedToken, err := token.SignedString([]byte(testSecret))
				if err != nil {
					t.Fatalf("Failed to create token: %v", err)
				}
				return signedToken
			},
			wantRole: "",
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tokenString := tc.tokenFunc(t)

			role, err := jwtpkg.GetRoleFromToken(tokenString)

			if (err != nil) != tc.wantErr {
				t.Errorf("GetRoleFromToken() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if role != tc.wantRole {
				t.Errorf("GetRoleFromToken() role = %v, want %v", role, tc.wantRole)
			}
		})
	}
}

func TestIsAdmin(t *testing.T) {
	originalKey := jwtpkg.GetSecretKey()
	defer jwtpkg.SetSecretKey(originalKey)

	jwtpkg.SetSecretKey([]byte(testSecret))

	tests := []struct {
		name      string
		tokenFunc func(t *testing.T) string
		wantAdmin bool
		wantErr   bool
	}{
		{
			name: "Valid admin token",
			tokenFunc: func(t *testing.T) string {
				return createTestToken(t, testAdminUser, testAdminRole, time.Hour)
			},
			wantAdmin: true,
			wantErr:   false,
		},
		{
			name: "Valid user token",
			tokenFunc: func(t *testing.T) string {
				return createTestToken(t, testUsername, testRole, time.Hour)
			},
			wantAdmin: false,
			wantErr:   false,
		},
		{
			name: "Expired token",
			tokenFunc: func(t *testing.T) string {
				return createTestToken(t, testAdminUser, testAdminRole, -time.Second)
			},
			wantAdmin: false,
			wantErr:   true,
		},
		{
			name: "Invalid token",
			tokenFunc: func(t *testing.T) string {
				return "invalid.token"
			},
			wantAdmin: false,
			wantErr:   true,
		},
		{
			name: "Token without role",
			tokenFunc: func(t *testing.T) string {
				claims := jwt.MapClaims{
					"username": testUsername,
					"exp":      time.Now().Add(time.Hour).Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				signedToken, err := token.SignedString([]byte(testSecret))
				if err != nil {
					t.Fatalf("Failed to create token: %v", err)
				}
				return signedToken
			},
			wantAdmin: false,
			wantErr:   true,
		},
		{
			name: "BookKeeper role token",
			tokenFunc: func(t *testing.T) string {
				return createTestToken(t, testUsername, "BookKeeper", time.Hour)
			},
			wantAdmin: false,
			wantErr:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tokenString := tc.tokenFunc(t)

			isAdmin, err := jwtpkg.IsAdmin(tokenString)

			if (err != nil) != tc.wantErr {
				t.Errorf("IsAdmin() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if isAdmin != tc.wantAdmin {
				t.Errorf("IsAdmin() = %v, want %v", isAdmin, tc.wantAdmin)
			}
		})
	}
}

func TestSetAndGetSecretKey(t *testing.T) {
	originalKey := jwtpkg.GetSecretKey()
	defer jwtpkg.SetSecretKey(originalKey)

	newKey := []byte("new-test-secret-key")
	jwtpkg.SetSecretKey(newKey)

	retrievedKey := jwtpkg.GetSecretKey()

	if string(retrievedKey) != string(newKey) {
		t.Errorf("GetSecretKey() = %v, want %v", string(retrievedKey), string(newKey))
	}
}

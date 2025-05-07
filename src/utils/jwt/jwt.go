package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"tick_test/types"
)

type Claims struct {
	Username string
	Role     types.Role
}

var secretKey []byte

func SetSecretKey(key []byte) {
	secretKey = key
}

func GetSecretKey() []byte {
	return secretKey
}

func GenerateToken(username string, role types.Role, duration time.Duration) (string, error) {
	if username == "" {
		return "", errors.New("username cannot be empty")
	}

	if role == "" {
		return "", errors.New("role cannot be empty")
	}

	if secretKey == nil {
		return "", errors.New("secret key is not set")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"username": username,
		"role":     role,
		"exp":      now.Add(duration).Unix(),
		"iat":      now.Unix(),
		"issuer":   "tick_test",
		"subject":  username,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ValidateToken(tokenString string) (Claims, error) {
	var claims Claims

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the algorithm is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return claims, err
	}

	if !token.Valid {
		return claims, errors.New("invalid token")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return claims, errors.New("invalid token claims")
	}

	username, ok := mapClaims["username"].(string)
	if !ok {
		return claims, errors.New("missing or invalid username claim")
	}

	role, ok := mapClaims["role"].(string)
	if !ok {
		return claims, errors.New("missing or invalid role claim")
	}

	claims.Username = username
	claims.Role = types.Role(role)

	return claims, nil
}

func GetUserFromToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	username, ok := mapClaims["username"].(string)
	if !ok {
		return "", errors.New("missing or invalid username claim")
	}

	return username, nil
}

func GetRoleFromToken(tokenString string) (types.Role, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	strRole, ok := mapClaims["role"].(string)
	role := types.Role(strRole)
	if !ok {
		return "", errors.New("missing or invalid role claim")
	}

	return role, nil
}

func IsAdmin(tokenString string) (bool, error) {
	role, err := GetRoleFromToken(tokenString)
	if err != nil {
		return false, err
	}

	return role == "Admin", nil
}

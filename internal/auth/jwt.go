package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey []byte

func init() {
	key := os.Getenv("JWT_SECRET")
	if key == "" {
		key = "cloud-mcp-secret-key-change-in-production"
	}
	secretKey = []byte(key)
}

// GenerateToken 生成JWT Token
func GenerateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString(secretKey)
}

// ParseToken 解析JWT Token
func ParseToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", jwt.ErrSignatureInvalid
	}

	username, _ := claims["username"].(string)
	return username, nil
}

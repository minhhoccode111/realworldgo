package utils

import (
	"fmt"
	"log"
	"time"

	"github.com/minhhoccode111/realworldgo/internal/config"
	"github.com/minhhoccode111/realworldgo/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(jwtConfig config.JWTConfig, user *model.User) (string, error) {
	if jwtConfig.Secret == "" {
		return "", fmt.Errorf("JWT secret cannot be empty")
	}
	secretKey := []byte(jwtConfig.Secret)
	expirationTime := time.Now().Add(jwtConfig.Expiration).Unix() // Calculate future expiration
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": user.Id,
		"exp":    expirationTime, // Use the calculated Unix timestamp
		"iat":    time.Now().Unix(),
		"iss":    jwtConfig.Issuer,
	})
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		log.Printf("error signing token: %v", err)
		return "", fmt.Errorf("error signing token: %v", err)
	}
	return tokenString, nil
}

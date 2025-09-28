package utils

import (
	"log"
	"time"

	"github.com/minhhoccode111/realworldgo/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// TODO: use only userId is insecure
func GenerateJWT(jwtConfig config.JWTConfig, userId string) (string, error) {
	secretKey := []byte(jwtConfig.Secret)
	expirationTime := time.Now().Add(jwtConfig.Expiration).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId,
		"exp":    expirationTime,
		"iat":    time.Now().Unix(),
		"iss":    jwtConfig.Issuer,
	})
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		log.Printf("error signing token: %v", err)
		return "", err
	}
	return tokenString, nil
}

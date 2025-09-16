package utils

import (
	"auth/internal/config"
	"auth/internal/model"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func UserToUserDTO(user *model.User) model.UserDTO {
	return model.UserDTO{
		Id:       user.Id,
		Email:    user.Email,
		IsActive: user.IsActive,
		Role:     user.Role,
	}
}

func GenerateJWT(jwtConfig config.JWTConfig, user *model.UserDTO) (string, error) {
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

func ValidatePassword(hashedPassword string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

func HashedPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func IsValidEmail(email string) (string, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return "", fmt.Errorf("email cannot be empty")
	}
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !re.MatchString(email) {
		return "", fmt.Errorf("invalid email: %v", email)
	}
	return email, nil
}

func IsValidPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if len(password) < 8 {
		return "", fmt.Errorf("password must be at least 8 characters long")
	}
	hasUpper, hasLower, hasDigit, hasSpecial := false, false, false, false
	for _, ch := range password {
		if unicode.IsUpper(ch) {
			hasUpper = true
		} else if unicode.IsLower(ch) {
			hasLower = true
		} else if unicode.IsDigit(ch) {
			hasDigit = true
		} else if strings.ContainsAny(string(ch), "!@#$%^&*()_+{}|:\"<>?~") {
			hasSpecial = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return "", fmt.Errorf(
			"weak password: '%v'. Must contain at least: 1 uppercase, 1 lowercase, 1 digit, 1 special character",
			password,
		)
	}
	return password, nil
}

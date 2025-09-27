package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

func IsValidUsername(username string) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return "", fmt.Errorf("username cannot be empty")
	}
	re := regexp.MustCompile(`^[a-zA-Z0-9]{2,50}$`)
	if !re.MatchString(username) {
		return "", fmt.Errorf("invalid username: %v", username)
	}
	return username, nil
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

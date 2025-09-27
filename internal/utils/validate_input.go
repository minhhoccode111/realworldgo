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
	re := regexp.MustCompile(`^[a-zA-Z0-9]{2,}$`)
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
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{5,}$`)
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

func IsNotEmpty(str string) (string, error) {
	str = strings.TrimSpace(str)
	if str == "" {
		return "", fmt.Errorf("string cannot be empty")
	}
	return str, nil
}

func IsValidTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", fmt.Errorf("title cannot be empty")
	}
	return title, nil
}

func IsValidBody(body string) (string, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return "", fmt.Errorf("body cannot be empty")
	}
	return body, nil
}

func IsValidDescription(description string) (string, error) {
	description = strings.TrimSpace(description)
	if description == "" {
		return "", fmt.Errorf("description cannot be empty")
	}
	return description, nil
}

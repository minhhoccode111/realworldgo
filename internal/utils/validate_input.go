package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

func IsValidUsername(username string) (string, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return "", fmt.Errorf("username cannot be empty")
	}
	re := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !re.MatchString(username) {
		return "", fmt.Errorf("username must contain only letters and numbers (a-z, A-Z, 0-9)")
	}
	if l := utf8.RuneCountInString(username); l < 2 || l > 50 {
		return "", fmt.Errorf("username must be between 2 and 50 characters")
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
		return "", fmt.Errorf("invalid email format")
	}
	if l := utf8.RuneCountInString(email); l < 5 || l > 320 {
		return "", fmt.Errorf("email must be between 5 and 320 characters")
	}
	return email, nil
}

func IsValidPassword(password string) (string, error) {
	password = strings.TrimSpace(password)
	if l := utf8.RuneCountInString(password); l < 8 || l > 50 {
		return "", fmt.Errorf("password must be between 8 and 50 characters")
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

func IsValidTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", fmt.Errorf("title cannot be empty")
	}
	if utf8.RuneCountInString(title) > 255 {
		return "", fmt.Errorf("title exceeds 255 characters")
	}
	return title, nil
}

func IsValidDescription(description string) (string, error) {
	description = strings.TrimSpace(description)
	if description == "" {
		return "", fmt.Errorf("description cannot be empty")
	}
	if utf8.RuneCountInString(description) > 255 {
		return "", fmt.Errorf("description exceeds 255 characters")
	}
	return description, nil
}

func IsValidTagList(tagList []string) ([]string, error) {
	tags := []string{}
	for _, tag := range tagList {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return nil, fmt.Errorf("A tag cannot be empty")
		}
		if utf8.RuneCountInString(tag) > 50 {
			return nil, fmt.Errorf("A tag exceeds 50 characters")
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

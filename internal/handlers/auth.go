package handlers

import "auth/internal/services"

type AuthHandler struct {
	AuthService services.AuthService
}

package middleware

import (
	"auth/internal/services"
	"net/http"
)

func AuthMiddleware(authService services.AuthService) func(http.Handler) http.Handler

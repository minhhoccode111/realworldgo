package middleware

import (
	"auth/internal/config"
	"net/http"
)

func CORSMiddleware(config *config.Config) func(http.Handler) http.Handler

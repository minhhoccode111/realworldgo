package middleware

import "net/http"

func LoggingMiddleware() func(http.Handler) http.Handler

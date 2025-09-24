package middleware

import (
	"context"
	"net/http"
	"time"
)

// timeout middleware use context
func TimeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// WARN: only allow the request to be ran in 100 milliseconds, just
		// playing around with context this earlier that WriteTimeout to make
		// sure handler finishes before server kill connection
		ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*7500)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

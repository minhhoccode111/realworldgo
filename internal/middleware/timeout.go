package middleware

import (
	"context"
	"net/http"
	"time"
)

// timeout middleware use context
func TimeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// INFO: this is not the same as ReadTimeout or WriteTimeout
		ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*7500)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

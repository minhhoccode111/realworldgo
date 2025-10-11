package middleware

import (
	"net/http"
	"slices"
)

// CORS middleware
func CorsMiddleware(AllowedOrigins []string) func(http.Handler) http.Handler {
	allowAll := slices.Contains(AllowedOrigins, "*")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			var allowed bool
			if allowAll {
				allowed = true
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if slices.Contains(AllowedOrigins, origin) {
				allowed = true
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if allowed {
				w.Header().
					Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
				w.Header().
					Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
				// w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
			}

			if r.Method == http.MethodOptions {
				if allowed {
					w.WriteHeader(http.StatusNoContent)
				} else {
					w.WriteHeader(http.StatusForbidden)
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

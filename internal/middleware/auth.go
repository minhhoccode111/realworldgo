package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/minhhoccode111/realworldgo/internal/database"
	. "github.com/minhhoccode111/realworldgo/internal/utils"
)

// Authentication middleware
// TODO: make this authentication middleware allows optional param
func AuthMiddleware(jwtSecret string, db database.Service) func(http.HandlerFunc) http.HandlerFunc {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// get authentication header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				WriteJSON(w, http.StatusUnauthorized, JSON{"error": "no authorization header"})
				return
			}
			parts := strings.Fields(authHeader)
			if !strings.EqualFold(parts[0], "Token") {
				WriteJSON(w, http.StatusUnauthorized,
					JSON{"error": "authorization header must start with 'Token'"},
				)
				return
			}
			if len(parts) != 2 {
				WriteJSON(w, http.StatusUnauthorized,
					JSON{"error": "authorization header must be formatted as 'Token <token>'"},
				)
				return
			}
			tokenStr := parts[1]
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})
			if err != nil {
				log.Printf("Error parsing token: %v", err)
				WriteJSON(w, http.StatusUnauthorized, JSON{"error": err.Error()})
				return
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || !token.Valid {
				WriteJSON(w, http.StatusUnauthorized, JSON{"error": "invalid token"})
				return
			}
			userId, ok := claims[string(CtxUserIdKey)].(string)
			if !ok || userId == "" {
				WriteJSON(w, http.StatusUnauthorized, JSON{"error": "missing userId in token"})
				return
			}
			// pass context to database query
			user, err := db.SelectUserById(r.Context(), userId)
			if err != nil {
				log.Printf("Error selecting user by id: %v", err)
				if strings.Contains(err.Error(), "timeout") {
					WriteJSON(w, http.StatusUnauthorized, JSON{"error": err.Error()})
					return
				}
				WriteJSON(w, http.StatusUnauthorized, JSON{"error": "cannot authorize user in jwt"})
				return
			}
			ctx := context.WithValue(r.Context(), CtxUserKey, *user)
			handlerFunc(w, r.WithContext(ctx))
		}
	}
}

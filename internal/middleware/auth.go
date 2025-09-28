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
func AuthMiddleware(
	isOptional bool,
	jwtSecret string,
	db database.Service,
) func(http.HandlerFunc) http.HandlerFunc {
	return func(hf http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// first set isAuth to false
			ctx := context.WithValue(r.Context(), CtxIsAuthKey, false)

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// in each failure, check if auth is optional, continue if it is
				if isOptional {
					hf(w, r.WithContext(ctx))
					return
				}
				// else return immediately
				WriteJSON(w, http.StatusUnauthorized, JSON{"error": "no authorization header"})
				return
			}

			parts := strings.Fields(authHeader)
			if !strings.EqualFold(parts[0], "Token") {
				if isOptional {
					hf(w, r.WithContext(ctx))
					return
				}
				WriteJSON(w, http.StatusUnauthorized,
					JSON{"error": "authorization header must start with 'Token'"},
				)
				return
			}
			if len(parts) != 2 {
				if isOptional {
					hf(w, r.WithContext(ctx))
					return
				}
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
				if isOptional {
					hf(w, r.WithContext(ctx))
					return
				}
				log.Printf("Error parsing token: %v", err)
				WriteJSON(w, http.StatusUnauthorized, JSON{"error": err.Error()})
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok || !token.Valid {
				if isOptional {
					hf(w, r.WithContext(ctx))
					return
				}
				WriteJSON(w, http.StatusUnauthorized, JSON{"error": "invalid token"})
				return
			}

			userId, ok := claims[string(CtxUserIdKey)].(string)
			if !ok || userId == "" {
				if isOptional {
					hf(w, r.WithContext(ctx))
					return
				}
				WriteJSON(w, http.StatusUnauthorized, JSON{"error": "missing userId in token"})
				return
			}

			// pass request's context to database query
			user, err := db.SelectUserById(r.Context(), userId)
			if err != nil {
				if isOptional {
					hf(w, r.WithContext(ctx))
					return
				}
				log.Printf("Error selecting user by id: %v", err)
				WriteJSON(w, http.StatusUnauthorized, JSON{"error": err.Error()})
				return
			}

			// at this state, the user is authenticated
			ctx = context.WithValue(ctx, CtxIsAuthKey, true)
			ctx = context.WithValue(ctx, CtxUserKey, *user)
			hf(w, r.WithContext(ctx))
		}
	}
}

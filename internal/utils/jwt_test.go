package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/minhhoccode111/realworldgo/internal/config"
)

func TestGenerateJWT(t *testing.T) {
	cfg := config.JWTConfig{
		Secret:     "secret",
		Expiration: time.Minute,
		Issuer:     "tester",
	}

	tokenStr, err := GenerateJWT(cfg, "user-123")
	if err != nil {
		t.Fatalf("failed generating token: %v", err)
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		t.Fatalf("failed parsing token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		t.Fatalf("expected valid token claims")
	}

	if claims["userId"] != "user-123" {
		t.Fatalf("unexpected userId claim: %v", claims["userId"])
	}
	if claims["iss"] != "tester" {
		t.Fatalf("unexpected issuer claim: %v", claims["iss"])
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Fatalf("exp claim missing")
	}
	if time.Until(time.Unix(int64(exp), 0)) <= 0 {
		t.Fatalf("token expiration should be in the future")
	}
}

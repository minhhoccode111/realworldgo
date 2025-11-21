package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/minhhoccode111/realworldgo/internal/config"
	"github.com/minhhoccode111/realworldgo/internal/model"
	"github.com/minhhoccode111/realworldgo/internal/utils"
)

type stubUserLookup struct {
	user *model.User
	err  error
}

func (s stubUserLookup) SelectUser(ctx context.Context, id, email, username string) (*model.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.user, nil
}

func TestAuthMiddlewareOptionalSkipsAuth(t *testing.T) {
	called := false
	handler := AuthMiddleware(true, "secret", stubUserLookup{})(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)
	if !called {
		t.Fatalf("expected downstream handler to be invoked")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
}

func TestAuthMiddlewareRequiresHeader(t *testing.T) {
	handler := AuthMiddleware(false, "secret", stubUserLookup{})(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when authorization header missing, got %d", rec.Code)
	}
}

func TestAuthMiddlewareValidToken(t *testing.T) {
	const secret = "secret"
	user := &model.User{Id: "user-1", Email: "user@example.com"}

	token, err := utils.GenerateJWT(config.JWTConfig{
		Secret:     secret,
		Expiration: time.Hour,
		Issuer:     "test-suite",
	}, user.Id)
	if err != nil {
		t.Fatalf("failed to generate jwt: %v", err)
	}

	handler := AuthMiddleware(false, secret, stubUserLookup{user: user})(func(w http.ResponseWriter, r *http.Request) {
		isAuth, _ := r.Context().Value(utils.CtxIsAuthKey).(bool)
		if !isAuth {
			t.Fatalf("expected request context to be marked as authenticated")
		}

		gotUser, ok := r.Context().Value(utils.CtxUserKey).(model.User)
		if !ok || gotUser.Id != user.Id {
			t.Fatalf("expected user in context, got %#v", gotUser)
		}

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token "+token)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected handler to run successfully, got %d", rec.Code)
	}
}

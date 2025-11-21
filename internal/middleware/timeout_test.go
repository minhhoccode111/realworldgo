package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTimeoutMiddleware(t *testing.T) {
	done := make(chan struct{})
	handler := TimeoutMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(done)
		deadline, ok := r.Context().Deadline()
		if !ok {
			t.Fatalf("expected context deadline to be set")
		}
		timeout := time.Until(deadline)
		if timeout < 7*time.Second || timeout > 8*time.Second {
			t.Fatalf("expected about 7.5s timeout, got %s", timeout)
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("handler did not complete")
	}
}

package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteText(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteText(rec, http.StatusAccepted, "ok")

	resp := rec.Result()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/text" {
		t.Fatalf("unexpected content type: %s", ct)
	}

	body := rec.Body.String()
	if body != "ok" {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	payload := map[string]string{"message": "hello"}

	WriteJSON(rec, http.StatusCreated, payload)
	resp := rec.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content type: %s", ct)
	}

	var decoded map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	if decoded["message"] != "hello" {
		t.Fatalf("unexpected payload: %v", decoded)
	}
}

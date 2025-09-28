package utils

import (
	"encoding/json"
	"net/http"
)

type JSON map[string]any
type ctxKey string

const (
	CtxUserKey   ctxKey = "user"
	CtxUserIdKey ctxKey = "userId"
	CtxIsAuthKey ctxKey = "isAuth"
)

func WriteText(w http.ResponseWriter, status int, data string) {
	w.Header().Set("Content-Type", "application/text")
	w.WriteHeader(status)
	w.Write([]byte(data))
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

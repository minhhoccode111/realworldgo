package utils

import (
	"net/http"
	"strconv"
	"strings"
)

// TODO: tag can be a slice of string
func SearchQueries(w http.ResponseWriter, r *http.Request) (string, string, string, int, int) {
	tag := strings.TrimSpace(r.URL.Query().Get("tag"))
	author := strings.TrimSpace(r.URL.Query().Get("author"))
	favorited := strings.TrimSpace(r.URL.Query().Get("favorited"))

	offsetStr := strings.TrimSpace(r.URL.Query().Get("offset"))
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	limitStr := strings.TrimSpace(r.URL.Query().Get("limit"))
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	return tag, author, favorited, limit, offset
}

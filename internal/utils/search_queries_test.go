package utils

import (
	"net/http"
	"net/url"
	"testing"
)

func TestSearchQueries(t *testing.T) {
	values := url.Values{}
	values.Set("tag", " go ")
	values.Set("author", "alice")
	values.Set("favorited", "bob")
	values.Set("limit", "50")
	values.Set("offset", "10")

	req := &http.Request{URL: &url.URL{RawQuery: values.Encode()}}

	tag, author, favorited, limit, offset := SearchQueries(req)

	if tag != "go" || author != "alice" || favorited != "bob" {
		t.Fatalf("unexpected strings: %s %s %s", tag, author, favorited)
	}
	if limit != 50 || offset != 10 {
		t.Fatalf("unexpected pagination: %d %d", limit, offset)
	}
}

func TestSearchQueriesWithInvalidValues(t *testing.T) {
	values := url.Values{}
	values.Set("limit", "-1")
	values.Set("offset", "-10")

	req := &http.Request{URL: &url.URL{RawQuery: values.Encode()}}
	_, _, _, limit, offset := SearchQueries(req)

	if limit != 20 || offset != 0 {
		t.Fatalf("expected defaults for invalid inputs, got limit=%d offset=%d", limit, offset)
	}

	values.Set("limit", "500")
	req.URL.RawQuery = values.Encode()
	_, _, _, limit, _ = SearchQueries(req)
	if limit != 20 {
		t.Fatalf("expected limit to clamp to 20 when exceeding max, got %d", limit)
	}
}

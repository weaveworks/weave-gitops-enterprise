package commands

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alecthomas/assert"
)

func TestGetIssuer(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, `{"issuer": "https://example.com/issuer"}`)
	}))
	defer mockServer.Close()

	issuer, err := getIssuer(mockServer.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedIssuer := "https://example.com/issuer"
	assert.Equal(t, expectedIssuer, issuer, "Expected issuer %s, got %s", expectedIssuer, issuer)
}

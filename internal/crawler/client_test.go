package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alexandrmotologa/brokenlens/internal/config"
)

func TestHTTPClientAuthAndHeaders(t *testing.T) {
	var receivedAuth string
	var receivedCustomHeader string
	var receivedCookie string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		receivedCustomHeader = r.Header.Get("X-Staging-Token")
		receivedCookie = r.Header.Get("Cookie")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer ts.Close()

	cfg := &config.Config{
		TargetURL:   ts.URL,
		Timeout:     5 * time.Second,
		BasicAuth:   "admin:secretPass123",
		Headers:     map[string]string{"X-Staging-Token": "test-token-value"},
		Cookie:      "session_id=abc123xyz",
		Concurrency: 1,
	}

	client := NewHTTPClient(cfg)
	resp := client.Fetch(context.Background(), ts.URL, http.MethodGet, true)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}

	if receivedAuth == "" {
		t.Error("Expected Authorization header to be set via BasicAuth")
	}

	if receivedCustomHeader != "test-token-value" {
		t.Errorf("Expected X-Staging-Token 'test-token-value', got %q", receivedCustomHeader)
	}

	if receivedCookie != "session_id=abc123xyz" {
		t.Errorf("Expected Cookie 'session_id=abc123xyz', got %q", receivedCookie)
	}
}

func TestHTTPClientRetryAfter(t *testing.T) {
	var attempts atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := attempts.Add(1)
		if count == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte("Rate limit exceeded"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success on retry"))
	}))
	defer ts.Close()

	cfg := &config.Config{
		TargetURL: ts.URL,
		Timeout:   5 * time.Second,
	}

	client := NewHTTPClient(cfg)
	resp := client.Fetch(context.Background(), ts.URL, http.MethodGet, true)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 after retry, got %d", resp.StatusCode)
	}

	if attempts.Load() < 2 {
		t.Errorf("Expected at least 2 attempts due to 429 retry, got %d", attempts.Load())
	}
}

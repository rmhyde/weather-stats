package emoncms

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostTemperature(t *testing.T) {
	t.Run("json success response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", r.Method)
			}
			if r.URL.Path != "/input/post" {
				t.Errorf("path = %s, want /input/post", r.URL.Path)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm failed: %v", err)
			}
			if got := r.PostForm.Get("node"); got != "weather" {
				t.Errorf("node = %q, want %q", got, "weather")
			}
			if got := r.PostForm.Get("apikey"); got != "emon-secret" {
				t.Errorf("apikey = %q, want %q", got, "emon-secret")
			}
			var payload map[string]float64
			if err := json.Unmarshal([]byte(r.PostForm.Get("fulljson")), &payload); err != nil {
				t.Fatalf("invalid fulljson %q: %v", r.PostForm.Get("fulljson"), err)
			}
			if payload["temperature"] != 18.5 {
				t.Errorf("temperature = %v, want 18.5", payload["temperature"])
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true}`))
		}))
		defer srv.Close()

		client := NewClient(srv.URL, "emon-secret", srv.Client())
		if err := client.PostTemperature(context.Background(), "weather", "temperature", 18.5); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("legacy ok response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}))
		defer srv.Close()

		client := NewClient(srv.URL, "k", srv.Client())
		if err := client.PostTemperature(context.Background(), "1", "temperature", 12.3); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejected response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"success":false,"message":"Invalid API key"}`))
		}))
		defer srv.Close()

		client := NewClient(srv.URL, "bad", srv.Client())
		err := client.PostTemperature(context.Background(), "1", "temperature", 12.3)
		if err == nil || !strings.Contains(err.Error(), "Invalid API key") {
			t.Fatalf("expected rejection error, got: %v", err)
		}
	})
}

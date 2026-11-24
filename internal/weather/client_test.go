package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchTemperature(t *testing.T) {
	t.Run("success including zero degrees", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/current.json" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if got := r.URL.Query().Get("key"); got != "secret-weather-key" {
				t.Errorf("key = %q, want %q", got, "secret-weather-key")
			}
			if got := r.URL.Query().Get("q"); got != "SW1A 1AA" {
				t.Errorf("q = %q, want %q", got, "SW1A 1AA")
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"location":{"name":"London"},"current":{"temp_c":0.0}}`))
		}))
		defer srv.Close()

		client := NewClient(srv.URL, "secret-weather-key", srv.Client())
		res, err := client.FetchTemperature(context.Background(), "SW1A 1AA")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.TempC != 0.0 {
			t.Errorf("TempC = %v, want 0.0", res.TempC)
		}
		if res.LocationName != "London" {
			t.Errorf("LocationName = %q, want %q", res.LocationName, "London")
		}
	})

	t.Run("api error payload does not leak key", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"code":1006,"message":"No matching location found."}}`))
		}))
		defer srv.Close()

		client := NewClient(srv.URL, "secret-key", srv.Client())
		_, err := client.FetchTemperature(context.Background(), "INVALID")
		if err == nil || !strings.Contains(err.Error(), "No matching location found.") {
			t.Fatalf("expected API error message, got: %v", err)
		}
		if strings.Contains(err.Error(), "secret-key") {
			t.Errorf("error leaked API key: %v", err)
		}
	})

	t.Run("missing temp_c field", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"location":{"name":"London"},"current":{}}`))
		}))
		defer srv.Close()

		client := NewClient(srv.URL, "k", srv.Client())
		_, err := client.FetchTemperature(context.Background(), "SW1A 1AA")
		if err == nil {
			t.Fatal("expected error for missing temp_c, got nil")
		}
	})
}

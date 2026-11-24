package config

import (
	"strings"
	"testing"
	"time"
)

func envMap(m map[string]string) func(string) string {
	return func(k string) string {
		return m[k]
	}
}

func TestLoad(t *testing.T) {
	t.Run("defaults and trimming", func(t *testing.T) {
		cfg, err := Load(envMap(map[string]string{
			"WEATHER_API_KEY": "wkey",
			"EMONCMS_API_KEY": "ekey",
			"EMONCMS_NODE":    "weather",
			"POSTCODE":        " SW1A 1AA ",
			"EMONCMS_URL":     "https://emoncms.local/",
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Postcode != "SW1A 1AA" {
			t.Errorf("Postcode = %q, want %q", cfg.Postcode, "SW1A 1AA")
		}
		if cfg.EmoncmsURL != "https://emoncms.local" {
			t.Errorf("EmoncmsURL = %q, want %q", cfg.EmoncmsURL, "https://emoncms.local")
		}
		if cfg.EmoncmsInputName != DefaultEmoncmsInputName {
			t.Errorf("EmoncmsInputName = %q, want %q", cfg.EmoncmsInputName, DefaultEmoncmsInputName)
		}
		if cfg.PollInterval != DefaultPollInterval {
			t.Errorf("PollInterval = %v, want %v", cfg.PollInterval, DefaultPollInterval)
		}
	})

	t.Run("missing required env vars", func(t *testing.T) {
		_, err := Load(envMap(map[string]string{}))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		for _, want := range []string{"WEATHER_API_KEY", "POSTCODE", "EMONCMS_API_KEY", "EMONCMS_NODE"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q missing %q", err.Error(), want)
			}
		}
	})

	t.Run("custom interval and zero", func(t *testing.T) {
		cfg, err := Load(envMap(map[string]string{
			"WEATHER_API_KEY":    "wkey",
			"EMONCMS_API_KEY":    "ekey",
			"EMONCMS_NODE":       "10",
			"EMONCMS_INPUT_NAME": "outside_temp",
			"POSTCODE":           "EC1A 1BB",
			"POLL_INTERVAL":      "0",
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.PollInterval != 0 {
			t.Errorf("PollInterval = %v, want 0", cfg.PollInterval)
		}
		if cfg.EmoncmsInputName != "outside_temp" {
			t.Errorf("EmoncmsInputName = %q, want %q", cfg.EmoncmsInputName, "outside_temp")
		}
	})

	t.Run("negative interval rejected", func(t *testing.T) {
		_, err := Load(envMap(map[string]string{
			"WEATHER_API_KEY": "wkey",
			"EMONCMS_API_KEY": "ekey",
			"EMONCMS_NODE":    "10",
			"POSTCODE":        "EC1A 1BB",
			"POLL_INTERVAL":   "-5m",
		}))
		if err == nil {
			t.Fatal("expected error for negative POLL_INTERVAL, got nil")
		}
	})

	t.Run("custom valid interval", func(t *testing.T) {
		cfg, err := Load(envMap(map[string]string{
			"WEATHER_API_KEY": "wkey",
			"EMONCMS_API_KEY": "ekey",
			"EMONCMS_NODE":    "10",
			"POSTCODE":        "EC1A 1BB",
			"POLL_INTERVAL":   "5m",
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.PollInterval != 5*time.Minute {
			t.Errorf("PollInterval = %v, want %v", cfg.PollInterval, 5*time.Minute)
		}
	})
}

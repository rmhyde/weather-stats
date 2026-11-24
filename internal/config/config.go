package config

import (
	"fmt"
	"strings"
	"time"
)

const (
	DefaultWeatherAPIURL    = "https://api.weatherapi.com"
	DefaultEmoncmsURL       = "https://emoncms.org"
	DefaultEmoncmsInputName = "temperature"
	DefaultPollInterval     = 15 * time.Minute
)

// Config holds application settings loaded from environment variables.
type Config struct {
	WeatherAPIKey    string
	WeatherAPIURL    string
	Postcode         string
	EmoncmsAPIKey    string
	EmoncmsURL       string
	EmoncmsNode      string
	EmoncmsInputName string
	PollInterval     time.Duration
}

// Load reads and validates configuration using the provided environment lookup function.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		WeatherAPIKey:    strings.TrimSpace(getenv("WEATHER_API_KEY")),
		WeatherAPIURL:    strings.TrimSpace(getenv("WEATHER_API_URL")),
		Postcode:         strings.TrimSpace(getenv("POSTCODE")),
		EmoncmsAPIKey:    strings.TrimSpace(getenv("EMONCMS_API_KEY")),
		EmoncmsURL:       strings.TrimSpace(getenv("EMONCMS_URL")),
		EmoncmsNode:      strings.TrimSpace(getenv("EMONCMS_NODE")),
		EmoncmsInputName: strings.TrimSpace(getenv("EMONCMS_INPUT_NAME")),
		PollInterval:     DefaultPollInterval,
	}

	if cfg.WeatherAPIURL == "" {
		cfg.WeatherAPIURL = DefaultWeatherAPIURL
	}
	cfg.WeatherAPIURL = strings.TrimRight(cfg.WeatherAPIURL, "/")

	if cfg.EmoncmsURL == "" {
		cfg.EmoncmsURL = DefaultEmoncmsURL
	}
	cfg.EmoncmsURL = strings.TrimRight(cfg.EmoncmsURL, "/")

	if cfg.EmoncmsInputName == "" {
		cfg.EmoncmsInputName = DefaultEmoncmsInputName
	}

	if rawInterval := strings.TrimSpace(getenv("POLL_INTERVAL")); rawInterval != "" {
		d, err := time.ParseDuration(rawInterval)
		if err != nil {
			return Config{}, fmt.Errorf("invalid POLL_INTERVAL %q: %w", rawInterval, err)
		}
		if d < 0 {
			return Config{}, fmt.Errorf("invalid POLL_INTERVAL %q: must be >= 0", rawInterval)
		}
		cfg.PollInterval = d
	}

	var missing []string
	if cfg.WeatherAPIKey == "" {
		missing = append(missing, "WEATHER_API_KEY")
	}
	if cfg.Postcode == "" {
		missing = append(missing, "POSTCODE")
	}
	if cfg.EmoncmsAPIKey == "" {
		missing = append(missing, "EMONCMS_API_KEY")
	}
	if cfg.EmoncmsNode == "" {
		missing = append(missing, "EMONCMS_NODE")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

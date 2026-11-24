package weather

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Reading contains the temperature and resolved location from WeatherAPI.com.
type Reading struct {
	TempC        float64
	LocationName string
}

// Client queries WeatherAPI.com for current weather observations.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient constructs a WeatherAPI.com client.
func NewClient(baseURL, apiKey string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

type apiResponse struct {
	Location struct {
		Name string `json:"name"`
	} `json:"location"`
	Current struct {
		TempC *float64 `json:"temp_c"`
	} `json:"current"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// FetchTemperature queries WeatherAPI.com for the current temperature in Celsius at postcode.
func (c *Client) FetchTemperature(ctx context.Context, postcode string) (Reading, error) {
	u, err := url.Parse(c.baseURL + "/v1/current.json")
	if err != nil {
		return Reading{}, fmt.Errorf("invalid weather API URL: %w", err)
	}
	q := u.Query()
	q.Set("key", c.apiKey)
	q.Set("q", postcode)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Reading{}, fmt.Errorf("creating weather API request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Unwrap *url.Error so we do not leak the API key in logs.
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			return Reading{}, fmt.Errorf("weather API request failed: %w", urlErr.Err)
		}
		return Reading{}, fmt.Errorf("weather API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Reading{}, fmt.Errorf("reading weather API response: %w", err)
	}

	var parsed apiResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		if resp.StatusCode != http.StatusOK {
			return Reading{}, fmt.Errorf("weather API returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}
		return Reading{}, fmt.Errorf("decoding weather API response: %w", err)
	}

	if parsed.Error != nil {
		return Reading{}, fmt.Errorf("weather API error %d: %s", parsed.Error.Code, parsed.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return Reading{}, fmt.Errorf("weather API returned HTTP %d", resp.StatusCode)
	}
	if parsed.Current.TempC == nil {
		return Reading{}, errors.New("weather API response missing current.temp_c")
	}

	return Reading{
		TempC:        *parsed.Current.TempC,
		LocationName: parsed.Location.Name,
	}, nil
}

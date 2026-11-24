package emoncms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client posts sensor inputs to an Emoncms instance.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient constructs an Emoncms HTTP client.
func NewClient(baseURL, apiKey string, httpClient *http.Client) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

type apiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// PostTemperature sends a temperature value to Emoncms via /input/post.
func (c *Client) PostTemperature(ctx context.Context, node, inputName string, tempC float64) error {
	payload, err := json.Marshal(map[string]float64{inputName: tempC})
	if err != nil {
		return fmt.Errorf("encoding emoncms payload: %w", err)
	}

	form := url.Values{}
	form.Set("node", node)
	form.Set("fulljson", string(payload))
	form.Set("apikey", c.apiKey)

	endpoint := c.baseURL + "/input/post"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("creating emoncms request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("emoncms request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return fmt.Errorf("reading emoncms response: %w", err)
	}
	trimmed := strings.TrimSpace(string(body))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("emoncms returned HTTP %d: %s", resp.StatusCode, trimmed)
	}

	// Older Emoncms versions reply with plain "ok"; newer ones reply with JSON {"success":true}.
	if strings.EqualFold(trimmed, "ok") {
		return nil
	}

	var parsed apiResponse
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return fmt.Errorf("unexpected emoncms response %q: %w", trimmed, err)
	}
	if !parsed.Success {
		if parsed.Message != "" {
			return fmt.Errorf("emoncms rejected input: %s", parsed.Message)
		}
		return fmt.Errorf("emoncms rejected input: %s", trimmed)
	}

	return nil
}

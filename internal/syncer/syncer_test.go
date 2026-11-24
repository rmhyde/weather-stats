package syncer

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"weather-stats/internal/weather"
)

type stubFetcher struct {
	reading weather.Reading
	err     error
}

func (s *stubFetcher) FetchTemperature(_ context.Context, _ string) (weather.Reading, error) {
	return s.reading, s.err
}

type stubPoster struct {
	calls     atomic.Int32
	lastNode  string
	lastInput string
	lastTemp  float64
	err       error
}

func (s *stubPoster) PostTemperature(_ context.Context, node, inputName string, tempC float64) error {
	s.calls.Add(1)
	s.lastNode = node
	s.lastInput = inputName
	s.lastTemp = tempC
	return s.err
}

func TestSyncOnce(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("fetches and posts reading", func(t *testing.T) {
		fetcher := &stubFetcher{reading: weather.Reading{TempC: 14.2, LocationName: "Manchester"}}
		poster := &stubPoster{}
		s := New(fetcher, poster, Options{
			Postcode:  "M1 1AE",
			Node:      "weather",
			InputName: "temperature",
		}, logger)

		if err := s.Run(context.Background(), true); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if poster.lastNode != "weather" || poster.lastInput != "temperature" || poster.lastTemp != 14.2 {
			t.Errorf("got post (%q, %q, %v), want (%q, %q, %v)",
				poster.lastNode, poster.lastInput, poster.lastTemp, "weather", "temperature", 14.2)
		}
	})

	t.Run("propagates fetch error in once mode", func(t *testing.T) {
		fetcher := &stubFetcher{err: errors.New("boom")}
		poster := &stubPoster{}
		s := New(fetcher, poster, Options{Postcode: "M1 1AE", Node: "weather", InputName: "temperature"}, logger)

		if err := s.Run(context.Background(), true); err == nil {
			t.Fatal("expected error, got nil")
		}
		if poster.calls.Load() != 0 {
			t.Errorf("expected 0 post calls after fetch error, got %d", poster.calls.Load())
		}
	})

	t.Run("loops until context cancelled", func(t *testing.T) {
		fetcher := &stubFetcher{reading: weather.Reading{TempC: 10.0, LocationName: "London"}}
		poster := &stubPoster{}
		s := New(fetcher, poster, Options{
			Postcode:     "SW1A 1AA",
			Node:         "weather",
			InputName:    "temperature",
			PollInterval: 10 * time.Millisecond,
		}, logger)

		ctx, cancel := context.WithTimeout(context.Background(), 35*time.Millisecond)
		defer cancel()

		if err := s.Run(ctx, false); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := poster.calls.Load(); got < 2 {
			t.Errorf("expected at least 2 polls before timeout, got %d", got)
		}
	})
}

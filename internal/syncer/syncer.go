package syncer

import (
	"context"
	"log/slog"
	"time"

	"weather-stats/internal/weather"
)

// TemperatureFetcher retrieves a current temperature reading for a location.
type TemperatureFetcher interface {
	FetchTemperature(ctx context.Context, postcode string) (weather.Reading, error)
}

// TemperaturePoster publishes a temperature reading to an external store.
type TemperaturePoster interface {
	PostTemperature(ctx context.Context, node, inputName string, tempC float64) error
}

// Options configures how the Syncer polls and publishes temperatures.
type Options struct {
	Postcode     string
	Node         string
	InputName    string
	PollInterval time.Duration
}

// Syncer fetches weather readings and pushes them to Emoncms.
type Syncer struct {
	fetcher TemperatureFetcher
	poster  TemperaturePoster
	opts    Options
	logger  *slog.Logger
}

// New constructs a Syncer.
func New(fetcher TemperatureFetcher, poster TemperaturePoster, opts Options, logger *slog.Logger) *Syncer {
	return &Syncer{
		fetcher: fetcher,
		poster:  poster,
		opts:    opts,
		logger:  logger,
	}
}

// SyncOnce fetches the current temperature once and posts it to Emoncms.
func (s *Syncer) SyncOnce(ctx context.Context) error {
	reading, err := s.fetcher.FetchTemperature(ctx, s.opts.Postcode)
	if err != nil {
		return err
	}

	if err := s.poster.PostTemperature(ctx, s.opts.Node, s.opts.InputName, reading.TempC); err != nil {
		return err
	}

	s.logger.Info("synced temperature to emoncms",
		"postcode", s.opts.Postcode,
		"location", reading.LocationName,
		"temp_c", reading.TempC,
		"node", s.opts.Node,
		"input", s.opts.InputName,
	)
	return nil
}

// Run executes SyncOnce immediately and, unless once is true or PollInterval is 0,
// repeats on PollInterval until ctx is cancelled.
func (s *Syncer) Run(ctx context.Context, once bool) error {
	if once || s.opts.PollInterval == 0 {
		return s.SyncOnce(ctx)
	}

	s.logger.Info("starting weather-stats loop",
		"postcode", s.opts.Postcode,
		"node", s.opts.Node,
		"input", s.opts.InputName,
		"interval", s.opts.PollInterval.String(),
	)

	if err := s.SyncOnce(ctx); err != nil {
		s.logger.Error("sync failed; will retry on next tick", "error", err)
	}

	ticker := time.NewTicker(s.opts.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("shutting down")
			return nil
		case <-ticker.C:
			if err := s.SyncOnce(ctx); err != nil {
				s.logger.Error("sync failed; will retry on next tick", "error", err)
			}
		}
	}
}

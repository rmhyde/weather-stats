package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"weather-stats/internal/config"
	"weather-stats/internal/emoncms"
	"weather-stats/internal/syncer"
	"weather-stats/internal/weather"
)

const httpTimeout = 15 * time.Second

func main() {
	once := flag.Bool("once", false, "fetch and push temperature once, then exit")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		logger.Error("configuration error", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpClient := &http.Client{Timeout: httpTimeout}
	weatherClient := weather.NewClient(cfg.WeatherAPIURL, cfg.WeatherAPIKey, httpClient)
	emoncmsClient := emoncms.NewClient(cfg.EmoncmsURL, cfg.EmoncmsAPIKey, httpClient)

	s := syncer.New(weatherClient, emoncmsClient, syncer.Options{
		Postcode:     cfg.Postcode,
		Node:         cfg.EmoncmsNode,
		InputName:    cfg.EmoncmsInputName,
		PollInterval: cfg.PollInterval,
	}, logger)

	if err := s.Run(ctx, *once); err != nil {
		logger.Error("execution failed", "error", err)
		os.Exit(1)
	}
}

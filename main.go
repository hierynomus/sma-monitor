package main

import (
	"context"
	"os"

	"github.com/hierynomus/sma-monitor/internal/config"
	"github.com/hierynomus/sma-monitor/internal/sma"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	iot "github.com/hierynomus/iot-monitor"
	"github.com/hierynomus/iot-monitor/pkg/monitor"
)

func main() {
	ctx := log.Logger.WithContext(context.Background())
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg := &config.Config{}
	boot := iot.NewBootstrapper("sma-monitor", "SMA solar monitor", "SMA", cfg, func() (*monitor.Monitor, error) {
		scraper, err := sma.NewScraper(ctx, cfg)
		if err != nil {
			return nil, err
		}

		return monitor.NewMonitor(ctx, cfg, scraper, sma.NewMetricProvider(cfg), sma.NewConverter(cfg))
	})

	boot.Binder.Cast("modbus.metrics", func(i interface{}) interface{} {
		b, err := yaml.Marshal(i)
		if err != nil {
			panic(err)
		}

		metrics := map[string]config.Metric{}

		if err := yaml.Unmarshal(b, metrics); err != nil {
			panic(err)
		}

		log.Logger.Info().Interface("metrics", metrics).Msg("Unmarshalled metrics")

		return metrics
	})

	if err := boot.Start(ctx); err != nil {
		log.Logger.Error().Err(err).Send()
		os.Exit(1)
	}
}

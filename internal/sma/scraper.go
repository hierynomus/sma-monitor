package sma

import (
	"context"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/grid-x/modbus"
	"github.com/hierynomus/iot-monitor/pkg/iot"
	"github.com/hierynomus/iot-monitor/pkg/logging"
	"github.com/hierynomus/sma-monitor/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var _ iot.Scraper = (*Scraper)(nil)

type Scraper struct {
	ctx     context.Context
	logger  zerolog.Logger
	cancel  context.CancelFunc
	handler *modbus.TCPClientHandler
	wg      sync.WaitGroup
	ch      chan iot.RawMessage
	ticker  *time.Ticker
	cfg     config.ModbusConfig
}

func NewScraper(ctx context.Context, cfg *config.Config) (*Scraper, error) {
	logger := logging.LoggerFor(ctx, "modbus")

	handler := modbus.NewTCPClientHandler(fmt.Sprintf("%s:%d", cfg.Modbus.Address, cfg.Modbus.Port))
	handler.Timeout = cfg.Modbus.Timeout
	handler.SlaveID = byte(cfg.Modbus.DeviceAddress)
	handler.Logger = &logger

	ctx, cancel := context.WithCancel(ctx)

	return &Scraper{
		ctx:     ctx,
		logger:  logging.LoggerFor(ctx, "sma-scraper"),
		cancel:  cancel,
		handler: handler,
		ch:      make(chan iot.RawMessage),
		wg:      sync.WaitGroup{},
		ticker:  time.NewTicker(cfg.Modbus.ScrapeInterval),
		cfg:     cfg.Modbus,
	}, nil
}

func (s *Scraper) Output() <-chan iot.RawMessage {
	return s.ch
}

func (s *Scraper) Stop() error {
	s.cancel()
	return nil
}

func (s *Scraper) Wait() {
	s.wg.Wait()
}

func (s *Scraper) Start(ctx context.Context) error {
	s.wg.Add(1)

	go s.run(ctx)

	return nil
}

func (s *Scraper) run(ctx context.Context) {
	ctx = s.logger.WithContext(ctx)

	defer s.wg.Done()
	defer close(s.ch)

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info().Msg("Stopping scraper")

			return
		case <-s.ticker.C:
			s.logger.Debug().Msg("Scraping SMA inverter")

			msg, err := s.scrape(ctx)
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to scrape SMA inverter")
				continue
			}

			s.logger.Debug().Interface("msg", msg).Msg("Scraped SMA inverter")

			yaml, err := yaml.Marshal(msg)
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to marshal message")
				continue
			}

			s.ch <- iot.RawMessage(yaml)
		}
	}
}

func (s *Scraper) scrape(ctx context.Context) (map[string]float64, error) {
	logger := logging.LoggerFor(ctx, "sma-scraper")

	if err := s.handler.Connect(); err != nil {
		return nil, err
	}

	defer s.handler.Close()

	client := modbus.NewClient(s.handler)

	msg := map[string]float64{}
	for name, metric := range s.cfg.Metrics {
		value, err := s.readMetric(ctx, client, name, metric)
		if err != nil {
			logger.Error().Str("metric", name).Err(err).Msg("Failed to read metric")
			continue
		}

		logger.Trace().Str("metric", name).Float64("value", value).Msg("Read metric")
		msg[name] = value
	}

	return msg, nil
}

func (s *Scraper) readMetric(ctx context.Context, client modbus.Client, name string, metric config.Metric) (float64, error) {
	logger := log.Ctx(ctx)

	value, err := client.ReadHoldingRegisters(metric.Modbus.Address, metric.Modbus.Size)
	if err != nil {
		return 0, err
	}

	logger.Trace().Uint16("address", metric.Modbus.Address).Str("metric", name).Bytes("value", value).Msg("Received SMA data")

	v := float64(binary.BigEndian.Uint32(value)) / float64(metric.Modbus.Scale)
	if v < metric.Modbus.Min || v > metric.Modbus.Max {
		return 0, fmt.Errorf("value %.2f out of range (%.2f,%.2f)", v, metric.Modbus.Min, metric.Modbus.Max)
	}

	return v, nil
}

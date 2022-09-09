package config

import (
	"time"

	"github.com/creasty/defaults"
	"github.com/hierynomus/iot-monitor/pkg/config"
)

var _ config.MonitorConfig = (*Config)(nil)

type Config struct {
	Modbus ModbusConfig      `yaml:"modbus" viper:"modbus" env:"MODBUS"`
	Http   config.HTTPConfig `yaml:"http" viper:"http"` //nolint:revive
}

func (c *Config) HTTP() config.HTTPConfig {
	return c.Http
}

func (c *Config) RawMessageContentType() string {
	return "application/json"
}

type ModbusConfig struct {
	ScrapeInterval time.Duration     `yaml:"scrape-interval" viper:"scrape-interval" env:"SCRAPE_INTERVAL"`
	Address        string            `yaml:"address" viper:"address"`
	Port           int               `yaml:"port" viper:"port" default:"502"`
	DeviceAddress  int               `yaml:"device-address" viper:"device-address" default:"3"` // 126 sunspec
	Timeout        time.Duration     `yaml:"timeout" viper:"timeout" default:"5s"`
	Namespace      string            `yaml:"namespace" viper:"namespace"`
	Metrics        map[string]Metric `yaml:"metrics" viper:"metrics"`
}

type Metric struct {
	Name string `yaml:"name"`
	Unit string `yaml:"unit"`
	// Type is one of: gauge, counter, histogram, summary
	Type string `yaml:"type"`
	// Help is the description of the metric
	Help string `yaml:"help"`
	// Labels are the labels of the metric
	Labels map[string]string `yaml:"labels"`
	Modbus ModBusMetric      `yaml:"modbus"`
}

type ModBusMetric struct {
	Address uint16  `yaml:"address"`
	Size    uint16  `yaml:"size" default:"2"`
	Scale   int     `yaml:"scale" default:"1"`
	Min     float64 `yaml:"min" default:"0"`
	Max     float64 `yaml:"max"`
}

func (m *ModBusMetric) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := defaults.Set(m); err != nil {
		return err
	}

	type n ModBusMetric

	if err := unmarshal((*n)(m)); err != nil {
		return err
	}

	return nil
}

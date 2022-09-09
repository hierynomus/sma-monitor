package sma

import (
	"fmt"

	"github.com/hierynomus/iot-monitor/pkg/iot"
	"github.com/hierynomus/sma-monitor/internal/config"
	"gopkg.in/yaml.v3"
)

type Converter struct {
	cfg *config.Config
}

var _ iot.Converter = (*Converter)(nil)

func NewConverter(cfg *config.Config) *Converter {
	return &Converter{
		cfg: cfg,
	}
}

func (c *Converter) Convert(metrics iot.RawMessage) (iot.MetricMessage, error) {
	m := map[string]float64{}

	if err := yaml.Unmarshal([]byte(metrics), &m); err != nil {
		return nil, err
	}

	msg := iot.MetricMessage{}

	for k, v := range m {
		msg[k] = iot.Metric{
			Value:    fmt.Sprintf("%f", v),
			Unit:     c.cfg.Modbus.Metrics[k].Unit,
			Absolute: true,
		}
	}

	return msg, nil
}

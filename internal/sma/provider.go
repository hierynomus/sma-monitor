package sma

import (
	"github.com/hierynomus/iot-monitor/pkg/iot"
	"github.com/hierynomus/sma-monitor/internal/config"
)

type Provider struct {
	metrics map[string]iot.MetricCollector
}

var _ iot.MetricProvider = (*Provider)(nil)

func NewMetricProvider(cfg *config.Config) *Provider {
	return &Provider{
		metrics: buildMetrics(cfg),
	}
}

func buildMetrics(cfg *config.Config) map[string]iot.MetricCollector {
	metrics := map[string]iot.MetricCollector{}
	for n, metric := range cfg.Modbus.Metrics {
		if metric.Type == "gauge" {
			metrics[n] = iot.NewGaugeL(cfg.Modbus.Namespace, metric.Name, metric.Help, metric.Labels)
		} else if metric.Type == "counter" {
			metrics[n] = iot.NewCounterL(cfg.Modbus.Namespace, metric.Name, metric.Help, metric.Labels)
		}
	}

	return metrics
}

func (p *Provider) Metrics() map[string]iot.MetricCollector {
	return p.metrics
}

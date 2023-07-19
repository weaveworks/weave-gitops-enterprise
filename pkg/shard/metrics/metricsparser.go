package metrics

import (
	"strings"

	"github.com/prometheus/common/expfmt"
)

type Metric struct {
	Name   string
	Help   string
	Type   string
	Labels map[string]string
	Value  float64
	Count  uint64
}

func parseMetrics(data []byte) ([]Metric, error) {
	parser := expfmt.TextParser{}
	metricFamilies, err := parser.TextToMetricFamilies(strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}

	var metrics []Metric

	for metricName, metricFamily := range metricFamilies {
		for _, metric := range metricFamily.GetMetric() {
			m := Metric{
				Name:   metricName,
				Help:   metricFamily.GetHelp(),
				Type:   metricFamily.GetType().String(),
				Labels: make(map[string]string),
			}

			for _, labelPair := range metric.GetLabel() {
				m.Labels[*labelPair.Name] = *labelPair.Value
			}

			switch {
			case metric.Gauge != nil:
				m.Value = metric.GetGauge().GetValue()
			case metric.Counter != nil:
				m.Value = metric.GetCounter().GetValue()
			case metric.Summary != nil:
				m.Value = metric.GetSummary().GetSampleSum()
				m.Count = metric.GetSummary().GetSampleCount()
			case metric.Histogram != nil:
				m.Value = metric.GetHistogram().GetSampleSum()
				m.Count = metric.GetHistogram().GetSampleCount()
			}

			metrics = append(metrics, m)
		}
	}

	return metrics, nil
}

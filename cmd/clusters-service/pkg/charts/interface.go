package charts

import "context"

// ChartClient implementations interact with Helm repositories.
type ChartClient interface {
	// ValuesForChart gets the key/value mapping parsed from the values.yaml.
	ValuesForChart(ctx context.Context, c *ChartReference) (map[string]interface{}, error)
	// FileFromChart gets the bytes for a named file from the chart.
	FileFromChart(ctx context.Context, c *ChartReference, filename string) ([]byte, error)
}

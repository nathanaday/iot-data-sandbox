package tools

import "fmt"

func init() {
	RegisterTool(
		ToolManifest{
			Name:          "MovingAverage",
			Description:   "Calculates the simple moving average of a dataset with a specified window size",
			Category:      CategoryAnalysis,
			Documentation: "MovingAverage.md",
			Parameters: []ParameterDefinition{
				{Name: "dataset", Type: "[]float64", Description: "The dataset to calculate moving average for", Required: true},
				{Name: "window_size", Type: "int", Description: "The number of points to include in each average calculation", Required: true},
			},
			Examples: []string{
				"Calculate 5-point moving average for smoothing noisy sensor data",
				"Apply a 24-hour moving average to hourly temperature readings",
			},
		},
		movingAverageWrapper,
	)
}

func movingAverageWrapper(params map[string]interface{}) (interface{}, error) {
	datasetRaw, ok := params["dataset"]
	if !ok {
		return nil, fmt.Errorf("missing required parameter: dataset")
	}

	windowSizeRaw, ok := params["window_size"]
	if !ok {
		return nil, fmt.Errorf("missing required parameter: window_size")
	}

	var dataset []float64
	switch v := datasetRaw.(type) {
	case []float64:
		dataset = v
	case []interface{}:
		dataset = make([]float64, len(v))
		for i, val := range v {
			f, ok := val.(float64)
			if !ok {
				return nil, fmt.Errorf("dataset[%d] is not a number", i)
			}
			dataset[i] = f
		}
	default:
		return nil, fmt.Errorf("dataset must be an array of numbers")
	}

	var windowSize int
	switch v := windowSizeRaw.(type) {
	case float64:
		windowSize = int(v)
	case int:
		windowSize = v
	default:
		return nil, fmt.Errorf("window_size must be an integer")
	}

	if windowSize < 1 {
		return nil, fmt.Errorf("window_size must be at least 1")
	}

	return movingAverage(dataset, windowSize), nil
}

// movingAverage calculates the simple moving average.
// For the first (windowSize-1) points, it uses all available prior points.
func movingAverage(dataset []float64, windowSize int) []float64 {
	n := len(dataset)
	if n == 0 {
		return []float64{}
	}

	result := make([]float64, n)

	for i := 0; i < n; i++ {
		start := i - windowSize + 1
		if start < 0 {
			start = 0
		}

		var sum float64
		count := 0
		for j := start; j <= i; j++ {
			sum += dataset[j]
			count++
		}
		result[i] = sum / float64(count)
	}

	return result
}

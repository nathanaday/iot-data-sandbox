package tools

import (
	"fmt"
	"sort"
)

func init() {
	RegisterTool(
		ToolManifest{
			Name:          "Smoothing",
			Description:   "Applies smoothing algorithms to reduce noise in a dataset. Supports multiple methods: simple moving average, exponential, and median filter.",
			Category:      CategoryFilter,
			Documentation: "Smoothing.md",
			Parameters: []ParameterDefinition{
				{Name: "dataset", Type: "[]float64", Description: "The dataset to smooth", Required: true},
				{Name: "method", Type: "string", Description: "Smoothing method: 'sma' (simple moving average), 'ema' (exponential), or 'median'", Required: false},
				{Name: "window_size", Type: "int", Description: "Window size for SMA and median methods (default: 3)", Required: false},
				{Name: "alpha", Type: "float64", Description: "Smoothing factor for EMA method, between 0 and 1 (default: 0.3)", Required: false},
			},
			Examples: []string{
				"Apply simple moving average smoothing with window size 5",
				"Use exponential smoothing with alpha=0.2 for gradual response",
				"Apply median filter to remove spike noise",
			},
		},
		smoothingWrapper,
	)
}

func smoothingWrapper(params map[string]interface{}) (interface{}, error) {
	datasetRaw, ok := params["dataset"]
	if !ok {
		return nil, fmt.Errorf("missing required parameter: dataset")
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

	method := "sma"
	if methodRaw, ok := params["method"]; ok {
		if m, ok := methodRaw.(string); ok {
			method = m
		} else {
			return nil, fmt.Errorf("method must be a string")
		}
	}

	windowSize := 3
	if wsRaw, ok := params["window_size"]; ok {
		switch v := wsRaw.(type) {
		case float64:
			windowSize = int(v)
		case int:
			windowSize = v
		default:
			return nil, fmt.Errorf("window_size must be an integer")
		}
	}

	alpha := 0.3
	if alphaRaw, ok := params["alpha"]; ok {
		switch v := alphaRaw.(type) {
		case float64:
			alpha = v
		case int:
			alpha = float64(v)
		default:
			return nil, fmt.Errorf("alpha must be a number")
		}
	}

	switch method {
	case "sma":
		if windowSize < 1 {
			return nil, fmt.Errorf("window_size must be at least 1")
		}
		return smoothSMA(dataset, windowSize), nil
	case "ema":
		if alpha <= 0 || alpha > 1 {
			return nil, fmt.Errorf("alpha must be between 0 (exclusive) and 1 (inclusive)")
		}
		return smoothEMA(dataset, alpha), nil
	case "median":
		if windowSize < 1 {
			return nil, fmt.Errorf("window_size must be at least 1")
		}
		return smoothMedian(dataset, windowSize), nil
	default:
		return nil, fmt.Errorf("unknown smoothing method: %s (use 'sma', 'ema', or 'median')", method)
	}
}

// smoothSMA applies simple moving average smoothing
func smoothSMA(dataset []float64, windowSize int) []float64 {
	n := len(dataset)
	if n == 0 {
		return []float64{}
	}

	result := make([]float64, n)

	for i := 0; i < n; i++ {
		start := i - windowSize/2
		end := i + windowSize/2

		if start < 0 {
			start = 0
		}
		if end >= n {
			end = n - 1
		}

		var sum float64
		count := 0
		for j := start; j <= end; j++ {
			sum += dataset[j]
			count++
		}
		result[i] = sum / float64(count)
	}

	return result
}

// smoothEMA applies exponential moving average smoothing
func smoothEMA(dataset []float64, alpha float64) []float64 {
	n := len(dataset)
	if n == 0 {
		return []float64{}
	}

	result := make([]float64, n)
	result[0] = dataset[0]

	for i := 1; i < n; i++ {
		result[i] = alpha*dataset[i] + (1-alpha)*result[i-1]
	}

	return result
}

// smoothMedian applies median filter smoothing
func smoothMedian(dataset []float64, windowSize int) []float64 {
	n := len(dataset)
	if n == 0 {
		return []float64{}
	}

	result := make([]float64, n)

	for i := 0; i < n; i++ {
		start := i - windowSize/2
		end := i + windowSize/2

		if start < 0 {
			start = 0
		}
		if end >= n {
			end = n - 1
		}

		window := make([]float64, end-start+1)
		copy(window, dataset[start:end+1])
		sort.Float64s(window)

		mid := len(window) / 2
		if len(window)%2 == 0 {
			result[i] = (window[mid-1] + window[mid]) / 2
		} else {
			result[i] = window[mid]
		}
	}

	return result
}

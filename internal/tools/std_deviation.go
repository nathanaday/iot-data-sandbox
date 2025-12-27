package tools

import (
	"fmt"
	"math"
)

func init() {
	RegisterTool(
		ToolManifest{
			Name:          "StandardDeviation",
			Description:   "Calculates the standard deviation of a dataset. Can compute overall or rolling standard deviation.",
			Category:      CategoryAnalysis,
			Documentation: "StandardDeviation.md",
			Parameters: []ParameterDefinition{
				{Name: "dataset", Type: "[]float64", Description: "The dataset to calculate standard deviation for", Required: true},
				{Name: "window_size", Type: "int", Description: "If provided, calculates rolling standard deviation with this window size. If omitted, returns overall standard deviation as a constant array.", Required: false},
			},
			Examples: []string{
				"Calculate overall standard deviation for quality control metrics",
				"Calculate 10-point rolling standard deviation to detect volatility changes",
			},
		},
		stdDeviationWrapper,
	)
}

func stdDeviationWrapper(params map[string]interface{}) (interface{}, error) {
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

	windowSizeRaw, hasWindow := params["window_size"]
	if !hasWindow {
		return stdDeviationOverall(dataset), nil
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

	if windowSize < 2 {
		return nil, fmt.Errorf("window_size must be at least 2 for standard deviation")
	}

	return stdDeviationRolling(dataset, windowSize), nil
}

// stdDeviationOverall calculates the population standard deviation and returns
// it as an array of the same length (constant value) for consistent output format.
func stdDeviationOverall(dataset []float64) []float64 {
	n := len(dataset)
	if n == 0 {
		return []float64{}
	}
	if n == 1 {
		return []float64{0}
	}

	var sum float64
	for _, v := range dataset {
		sum += v
	}
	mean := sum / float64(n)

	var sumSquaredDiff float64
	for _, v := range dataset {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}

	stdDev := math.Sqrt(sumSquaredDiff / float64(n))

	result := make([]float64, n)
	for i := range result {
		result[i] = stdDev
	}
	return result
}

// stdDeviationRolling calculates rolling standard deviation with the given window size.
func stdDeviationRolling(dataset []float64, windowSize int) []float64 {
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
		count := i - start + 1

		if count < 2 {
			result[i] = 0
			continue
		}

		var sum float64
		for j := start; j <= i; j++ {
			sum += dataset[j]
		}
		mean := sum / float64(count)

		var sumSquaredDiff float64
		for j := start; j <= i; j++ {
			diff := dataset[j] - mean
			sumSquaredDiff += diff * diff
		}

		result[i] = math.Sqrt(sumSquaredDiff / float64(count))
	}

	return result
}

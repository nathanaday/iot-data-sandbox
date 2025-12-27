package tools

import "fmt"

func init() {
	RegisterTool(
		ToolManifest{
			Name:          "CreateTrend",
			Description:   "Creates a trendline for a given dataset",
			Category:      CategoryAnalysis,
			Documentation: "CreateTrend.md",
			Parameters: []ParameterDefinition{
				{Name: "dataset", Type: "[]float64", Description: "The dataset to create a trendline for", Required: true},
			},
			Examples: []string{
				"Create a trendline for the dataset [1, 2, 3, 4, 5]",
			},
		},
		createTrendWrapper,
	)
}

// Wrapper that conforms to ToolFunc signature
func createTrendWrapper(params map[string]interface{}) (interface{}, error) {
	// Extract and validate parameters
	datasetRaw, ok := params["dataset"]
	if !ok {
		return nil, fmt.Errorf("missing required parameter: dataset")
	}

	// Type assertion - handle both []float64 and []interface{} (from JSON)
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

	return createTrend(dataset), nil
}

// Actual implementation with clean typed signature
// Uses linear regression (least squares method) to create a trendline
func createTrend(dataset []float64) []float64 {
	n := len(dataset)
	if n == 0 {
		return []float64{}
	}
	if n == 1 {
		return []float64{dataset[0]}
	}

	// Calculate sums for linear regression
	// y = mx + b where x is the index
	var sumX, sumY, sumXY, sumX2 float64
	for i, y := range dataset {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	nf := float64(n)
	denominator := nf*sumX2 - sumX*sumX

	// Handle edge case where all x values are the same (shouldn't happen with indices)
	if denominator == 0 {
		mean := sumY / nf
		result := make([]float64, n)
		for i := range result {
			result[i] = mean
		}
		return result
	}

	// Calculate slope and intercept
	slope := (nf*sumXY - sumX*sumY) / denominator
	intercept := (sumY - slope*sumX) / nf

	// Generate trendline values
	result := make([]float64, n)
	for i := range result {
		result[i] = slope*float64(i) + intercept
	}

	return result
}

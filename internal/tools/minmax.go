package tools

import (
	"fmt"
	"math"
)

func init() {
	RegisterTool(
		ToolManifest{
			Name:          "MinMax",
			Description:   "Finds the minimum and maximum values in a dataset along with their indices. Returns normalized data (0-1 scale) as output.",
			Category:      CategoryAnalysis,
			Documentation: "MinMax.md",
			Parameters: []ParameterDefinition{
				{Name: "dataset", Type: "[]float64", Description: "The dataset to analyze", Required: true},
			},
			Examples: []string{
				"Find peak and trough values in sensor readings",
				"Normalize data to 0-1 range for comparison across different scales",
			},
		},
		minMaxWrapper,
	)
}

func minMaxWrapper(params map[string]interface{}) (interface{}, error) {
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

	return minMax(dataset), nil
}

// MinMaxResult contains the analysis results
type MinMaxResult struct {
	Min        float64   `json:"min"`
	Max        float64   `json:"max"`
	MinIndex   int       `json:"min_index"`
	MaxIndex   int       `json:"max_index"`
	Range      float64   `json:"range"`
	Normalized []float64 `json:"normalized"`
}

// minMax analyzes the dataset and returns min/max info plus normalized values
func minMax(dataset []float64) MinMaxResult {
	n := len(dataset)
	if n == 0 {
		return MinMaxResult{
			Min:        0,
			Max:        0,
			MinIndex:   -1,
			MaxIndex:   -1,
			Range:      0,
			Normalized: []float64{},
		}
	}

	minVal := math.Inf(1)
	maxVal := math.Inf(-1)
	minIdx := 0
	maxIdx := 0

	for i, v := range dataset {
		if v < minVal {
			minVal = v
			minIdx = i
		}
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}

	rangeVal := maxVal - minVal

	normalized := make([]float64, n)
	if rangeVal == 0 {
		for i := range normalized {
			normalized[i] = 0.5
		}
	} else {
		for i, v := range dataset {
			normalized[i] = (v - minVal) / rangeVal
		}
	}

	return MinMaxResult{
		Min:        minVal,
		Max:        maxVal,
		MinIndex:   minIdx,
		MaxIndex:   maxIdx,
		Range:      rangeVal,
		Normalized: normalized,
	}
}

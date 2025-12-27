package tools

import "fmt"

func init() {
	RegisterTool(
		ToolManifest{
			Name:          "Derivative",
			Description:   "Calculates the rate of change (first derivative) between consecutive points in a dataset",
			Category:      CategoryTransform,
			Documentation: "Derivative.md",
			Parameters: []ParameterDefinition{
				{Name: "dataset", Type: "[]float64", Description: "The dataset to calculate derivatives for", Required: true},
			},
			Examples: []string{
				"Calculate velocity from position data",
				"Detect rapid changes in sensor readings",
				"Find acceleration from velocity measurements",
			},
		},
		derivativeWrapper,
	)
}

func derivativeWrapper(params map[string]interface{}) (interface{}, error) {
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

	return derivative(dataset), nil
}

// derivative calculates the first difference (discrete derivative) of the dataset.
// The output has the same length as input; the first value is 0 (no prior point to diff from).
func derivative(dataset []float64) []float64 {
	n := len(dataset)
	if n == 0 {
		return []float64{}
	}
	if n == 1 {
		return []float64{0}
	}

	result := make([]float64, n)
	result[0] = 0

	for i := 1; i < n; i++ {
		result[i] = dataset[i] - dataset[i-1]
	}

	return result
}

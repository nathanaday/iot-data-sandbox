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
func createTrend(dataset []float64) []float64 {
	// TODO: Implement trendline creation
	return dataset
}

package llm

import (
	"github.com/tmc/langchaingo/llms"
)

// GetAgentTools returns the LLM tools available to the agent for interacting
// with IoT data sandbox functionality.
func GetAgentTools() []llms.Tool {
	return []llms.Tool{
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "list_layers",
				Description: "List all data layers in the current project. Returns layer names, IDs, row counts, and time ranges. Use this to discover what data is available before performing analysis.",
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
					"required":   []string{},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "get_layer_data",
				Description: "Get time series data from a specific layer. Use this to analyze actual values in the dataset. Returns timestamps and values. For large datasets, use the 'sample_size' parameter to limit returned points.",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"layer_id": map[string]any{
							"type":        "integer",
							"description": "The ID of the layer to get data from",
						},
						"sample_size": map[string]any{
							"type":        "integer",
							"description": "Maximum number of data points to return. Use for large datasets to get a representative sample. Default is 100.",
						},
					},
					"required": []string{"layer_id"},
				},
			},
		},
		{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        "execute_analysis_tool",
				Description: "Execute an analysis tool on a data layer to create a new output layer with the results. Available tools: CreateTrend (creates a linear trend line), OutlierDetection (finds anomalies using z-score), MovingAverage (smooths data with rolling average), MinMax (normalizes values to 0-1 range), StandardDeviation (calculates rolling std dev), Smoothing (applies various smoothing algorithms), Derivative (calculates rate of change).",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"tool_name": map[string]any{
							"type":        "string",
							"description": "Name of the analysis tool to execute",
							"enum":        []string{"CreateTrend", "OutlierDetection", "MovingAverage", "MinMax", "StandardDeviation", "Smoothing", "Derivative"},
						},
						"source_layer_id": map[string]any{
							"type":        "integer",
							"description": "ID of the source layer to analyze",
						},
						"output_name": map[string]any{
							"type":        "string",
							"description": "Name for the new output layer that will be created with the analysis results",
						},
						"parameters": map[string]any{
							"type":        "object",
							"description": "Tool-specific parameters. For OutlierDetection: {threshold: number (default 2.0)}. For MovingAverage: {window_size: integer}. For Smoothing: {method: 'sma'|'ema'|'gaussian', window_size: integer, alpha: number for ema}. Other tools may not require additional parameters.",
						},
					},
					"required": []string{"tool_name", "source_layer_id", "output_name"},
				},
			},
		},
	}
}

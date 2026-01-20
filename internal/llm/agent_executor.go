package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
)

// ToolCallServiceInterface defines the interface for executing tools
type ToolCallServiceInterface interface {
	ExecuteToolOnLayer(toolName string, layerID int64, projectID int64, outputName string, params map[string]interface{}) (*ToolExecutionResult, error)
}

// DataLayerServiceInterface defines the interface for loading layer data
type DataLayerServiceInterface interface {
	LoadWithDataFrame(layerID int64) (*models.DataLayer, error)
}

// ToolExecutionResult represents the result of a tool execution
type ToolExecutionResult struct {
	ResultType string
	Message    string
	RowCount   int
	Layer      *models.DataLayer
}

// AgentExecutor handles execution of agent tool calls
type AgentExecutor struct {
	store            *persistence.Store
	toolCallService  ToolCallServiceInterface
	dataLayerService DataLayerServiceInterface
}

// NewAgentExecutor creates a new AgentExecutor
func NewAgentExecutor(store *persistence.Store, toolCallService ToolCallServiceInterface, dataLayerService DataLayerServiceInterface) *AgentExecutor {
	return &AgentExecutor{
		store:            store,
		toolCallService:  toolCallService,
		dataLayerService: dataLayerService,
	}
}

// ExecuteTool dispatches a tool call and returns the result as a JSON string
func (e *AgentExecutor) ExecuteTool(ctx context.Context, projectID int64, toolName string, arguments string) (string, error) {
	var args map[string]interface{}
	if arguments != "" && arguments != "{}" {
		if err := json.Unmarshal([]byte(arguments), &args); err != nil {
			return "", fmt.Errorf("failed to parse arguments: %w", err)
		}
	} else {
		args = make(map[string]interface{})
	}

	switch toolName {
	case "list_layers":
		return e.executeListLayers(ctx, projectID)
	case "get_layer_data":
		return e.executeGetLayerData(ctx, args)
	case "execute_analysis_tool":
		return e.executeAnalysisTool(ctx, projectID, args)
	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}

// LayerInfo represents layer information returned to the LLM
type LayerInfo struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	RowCount  int     `json:"row_count,omitempty"`
	StartTime *string `json:"start_time,omitempty"`
	EndTime   *string `json:"end_time,omitempty"`
	IsVisible bool    `json:"is_visible"`
}

// ListLayersResult is the result of the list_layers tool
type ListLayersResult struct {
	ProjectID   int64       `json:"project_id"`
	ProjectName string      `json:"project_name"`
	LayerCount  int         `json:"layer_count"`
	Layers      []LayerInfo `json:"layers"`
}

func (e *AgentExecutor) executeListLayers(ctx context.Context, projectID int64) (string, error) {
	// Load project
	project, err := e.store.LoadProject(projectID)
	if err != nil {
		return "", fmt.Errorf("failed to load project: %w", err)
	}

	// Load layers for project
	layers, err := e.store.LoadLayersByProjectId(projectID)
	if err != nil {
		return "", fmt.Errorf("failed to load layers: %w", err)
	}

	result := ListLayersResult{
		ProjectID:   projectID,
		ProjectName: project.Name,
		LayerCount:  len(layers),
		Layers:      make([]LayerInfo, 0, len(layers)),
	}

	for _, layer := range layers {
		info := LayerInfo{
			ID:        layer.DataLayerId,
			Name:      layer.Name,
			IsVisible: layer.IsVisible,
		}

		// Load metadata if layer has a dataframe
		if layer.DataFrameId != nil {
			_, dfSchema, err := e.store.LoadLayerWithDataFrame(layer.DataLayerId)
			if err == nil && dfSchema != nil {
				info.RowCount = dfSchema.RowCount
				if dfSchema.StartTime != nil {
					startStr := dfSchema.StartTime.Format(time.RFC3339)
					info.StartTime = &startStr
				}
				if dfSchema.EndTime != nil {
					endStr := dfSchema.EndTime.Format(time.RFC3339)
					info.EndTime = &endStr
				}
			}
		}

		result.Layers = append(result.Layers, info)
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonResult), nil
}

// DataPoint represents a single time series data point
type DataPoint struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

// GetLayerDataResult is the result of the get_layer_data tool
type GetLayerDataResult struct {
	LayerID    int64       `json:"layer_id"`
	LayerName  string      `json:"layer_name"`
	TotalRows  int         `json:"total_rows"`
	SampleSize int         `json:"sample_size"`
	Data       []DataPoint `json:"data"`
}

func (e *AgentExecutor) executeGetLayerData(ctx context.Context, args map[string]interface{}) (string, error) {
	// Extract layer_id
	layerIDRaw, ok := args["layer_id"]
	if !ok {
		return "", fmt.Errorf("layer_id is required")
	}
	layerID := int64(layerIDRaw.(float64))

	// Extract sample_size (default 100)
	sampleSize := 100
	if v, ok := args["sample_size"].(float64); ok {
		sampleSize = int(v)
	}

	// Load layer with data
	layer, err := e.dataLayerService.LoadWithDataFrame(layerID)
	if err != nil {
		return "", fmt.Errorf("failed to load layer: %w", err)
	}

	if layer.DataFrame == nil {
		result := map[string]interface{}{
			"error":      "layer has no data",
			"layer_id":   layerID,
			"layer_name": layer.Name,
		}
		jsonResult, _ := json.Marshal(result)
		return string(jsonResult), nil
	}

	// Extract data from DataFrame
	df := layer.DataFrame.Data
	timestamps := df.Col("timestamp").Records()

	// Find the value column (first non-timestamp column)
	var valueSeries []string
	for _, col := range df.Names() {
		if col != "timestamp" {
			valueSeries = df.Col(col).Records()
			break
		}
	}

	if valueSeries == nil {
		return "", fmt.Errorf("layer has no value column")
	}

	// Calculate total rows (skip header)
	totalRows := len(timestamps) - 1
	if totalRows < 0 {
		totalRows = 0
	}

	// Determine step for sampling
	step := 1
	if totalRows > sampleSize && sampleSize > 0 {
		step = totalRows / sampleSize
	}

	// Extract data points with sampling
	points := make([]DataPoint, 0)
	for i := 1; i < len(timestamps); i += step {
		if len(points) >= sampleSize {
			break
		}

		val, err := strconv.ParseFloat(valueSeries[i], 64)
		if err != nil {
			continue
		}

		points = append(points, DataPoint{
			Timestamp: timestamps[i],
			Value:     val,
		})
	}

	result := GetLayerDataResult{
		LayerID:    layerID,
		LayerName:  layer.Name,
		TotalRows:  totalRows,
		SampleSize: len(points),
		Data:       points,
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonResult), nil
}

// AnalysisToolResult is the result of the execute_analysis_tool tool
type AnalysisToolResult struct {
	Success      bool                   `json:"success"`
	ToolName     string                 `json:"tool_name"`
	ResultType   string                 `json:"result_type"`
	Message      string                 `json:"message"`
	RowCount     int                    `json:"row_count,omitempty"`
	CreatedLayer *CreatedLayerInfo      `json:"created_layer,omitempty"`
	Error        string                 `json:"error,omitempty"`
	RawResult    map[string]interface{} `json:"raw_result,omitempty"`
}

// CreatedLayerInfo contains info about a newly created layer
type CreatedLayerInfo struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (e *AgentExecutor) executeAnalysisTool(ctx context.Context, projectID int64, args map[string]interface{}) (string, error) {
	// Extract required parameters
	toolName, ok := args["tool_name"].(string)
	if !ok {
		return "", fmt.Errorf("tool_name is required")
	}

	sourceLayerIDRaw, ok := args["source_layer_id"]
	if !ok {
		return "", fmt.Errorf("source_layer_id is required")
	}
	sourceLayerID := int64(sourceLayerIDRaw.(float64))

	outputName, ok := args["output_name"].(string)
	if !ok {
		return "", fmt.Errorf("output_name is required")
	}

	// Extract optional parameters
	params := make(map[string]interface{})
	if p, ok := args["parameters"].(map[string]interface{}); ok {
		params = p
	}

	// Execute the tool
	toolExecResult, err := e.toolCallService.ExecuteToolOnLayer(
		toolName,
		sourceLayerID,
		projectID,
		outputName,
		params,
	)
	if err != nil {
		result := AnalysisToolResult{
			Success:  false,
			ToolName: toolName,
			Error:    err.Error(),
		}
		jsonResult, _ := json.Marshal(result)
		return string(jsonResult), nil
	}

	result := AnalysisToolResult{
		Success:    true,
		ToolName:   toolName,
		ResultType: toolExecResult.ResultType,
		Message:    toolExecResult.Message,
		RowCount:   toolExecResult.RowCount,
	}

	if toolExecResult.Layer != nil {
		result.CreatedLayer = &CreatedLayerInfo{
			ID:    toolExecResult.Layer.DataLayerId,
			Name:  toolExecResult.Layer.Name,
			Color: toolExecResult.Layer.Color,
		}
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	return string(jsonResult), nil
}

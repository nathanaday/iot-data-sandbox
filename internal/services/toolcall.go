package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/timeseries"
	"github.com/nathanaday/iot-data-sandbox/internal/tools"
)

type ToolCallService struct {
	store            *persistence.Store
	dataframeService *DataFrameService
	dataLayerService *DataLayerService
}

func NewToolCallService(store *persistence.Store) *ToolCallService {
	dataframeService := NewDataFrameService(store)
	dataLayerService := NewDataLayerService(store, dataframeService)

	return &ToolCallService{
		store:            store,
		dataframeService: dataframeService,
		dataLayerService: dataLayerService,
	}
}

func (s *ToolCallService) GetToolManifests() []tools.ToolManifest {
	return tools.GetAllToolManifests()
}

func (s *ToolCallService) CallTool(toolName string, parameters map[string]interface{}) (interface{}, error) {
	return tools.CallTool(toolName, parameters)
}

// CallToolJSON executes a tool and returns the result as a JSON string
func (s *ToolCallService) CallToolJSON(toolName string, parameters map[string]interface{}) (string, error) {
	result, err := tools.CallTool(toolName, parameters)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// ToolExecuteResult contains the result of executing a tool on a layer
type ToolExecuteResult struct {
	Layer       *models.DataLayer
	ResultType  string      // "array" for []float64 results, "object" for structured results
	RawResult   interface{} // The raw result from the tool
	RowCount    int
	Message     string
}

// ExecuteToolOnLayer executes a tool using data from a source layer and creates a new layer with the results
func (s *ToolCallService) ExecuteToolOnLayer(
	toolName string,
	sourceLayerId int64,
	projectId int64,
	outputName string,
	parameters map[string]interface{},
) (*ToolExecuteResult, error) {
	// 1. Load the source layer with its data
	sourceLayer, err := s.dataLayerService.LoadWithDataFrame(sourceLayerId)
	if err != nil {
		return nil, fmt.Errorf("failed to load source layer: %w", err)
	}

	if sourceLayer.DataFrame == nil {
		return nil, fmt.Errorf("source layer has no data")
	}

	// 2. Extract timestamps and values from the DataFrame
	df := sourceLayer.DataFrame.Data
	timestampSeries := df.Col("timestamp")
	timestamps := timestampSeries.Records()

	// Get first value column
	var valueSeries []string
	columnNames := df.Names()
	for _, colName := range columnNames {
		if colName != "timestamp" {
			valueSeries = df.Col(colName).Records()
			break
		}
	}

	if valueSeries == nil {
		return nil, fmt.Errorf("source layer has no value column")
	}

	// Convert to []float64 and []time.Time (skip header row)
	dataset := make([]float64, 0, len(valueSeries)-1)
	sourceTimestamps := make([]time.Time, 0, len(timestamps)-1)

	for i := 1; i < len(valueSeries); i++ {
		val, err := strconv.ParseFloat(valueSeries[i], 64)
		if err != nil {
			continue
		}
		dataset = append(dataset, val)

		ts, err := time.Parse(time.RFC3339, timestamps[i])
		if err != nil {
			continue
		}
		sourceTimestamps = append(sourceTimestamps, ts)
	}

	// 3. Inject dataset into parameters
	parameters["dataset"] = dataset

	// 4. Execute the tool
	result, err := tools.CallTool(toolName, parameters)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	// 5. Process result based on type
	executeResult := &ToolExecuteResult{
		RawResult: result,
	}

	// Check if result is an array that can be stored as a new layer
	switch v := result.(type) {
	case []float64:
		// Direct float64 array - create new layer
		layer, err := s.createLayerFromFloatArray(projectId, outputName, sourceTimestamps, v)
		if err != nil {
			return nil, fmt.Errorf("failed to create output layer: %w", err)
		}
		executeResult.Layer = layer
		executeResult.ResultType = "array"
		executeResult.RowCount = len(v)
		executeResult.Message = fmt.Sprintf("Created layer with %d data points", len(v))

	case tools.MinMaxResult:
		// MinMax result - use normalized values
		layer, err := s.createLayerFromFloatArray(projectId, outputName, sourceTimestamps, v.Normalized)
		if err != nil {
			return nil, fmt.Errorf("failed to create output layer: %w", err)
		}
		executeResult.Layer = layer
		executeResult.ResultType = "object"
		executeResult.RowCount = len(v.Normalized)
		executeResult.Message = fmt.Sprintf("Created layer with %d normalized data points (min=%.2f, max=%.2f)", len(v.Normalized), v.Min, v.Max)

	case tools.OutlierResult:
		// OutlierDetection result - extract only the outlier points with original values
		outlierTimestamps := make([]time.Time, len(v.OutlierIndices))
		outlierValues := make([]float64, len(v.OutlierIndices))
		for i, idx := range v.OutlierIndices {
			if idx < len(sourceTimestamps) && idx < len(dataset) {
				outlierTimestamps[i] = sourceTimestamps[idx]
				outlierValues[i] = dataset[idx]
			}
		}

		layer, err := s.createLayerFromFloatArray(projectId, outputName, outlierTimestamps, outlierValues)
		if err != nil {
			return nil, fmt.Errorf("failed to create output layer: %w", err)
		}
		executeResult.Layer = layer
		executeResult.ResultType = "object"
		executeResult.RowCount = len(v.OutlierIndices)
		executeResult.Message = fmt.Sprintf("Created layer with %d outlier points (threshold: %.1f std dev)", len(v.OutlierIndices), v.Threshold)

	default:
		// Unknown structured result - no layer created
		executeResult.ResultType = "unknown"
		executeResult.Message = "Tool returned non-array result (no layer created)"
	}

	return executeResult, nil
}

// createLayerFromFloatArray creates a new layer from timestamps and values
func (s *ToolCallService) createLayerFromFloatArray(
	projectId int64,
	layerName string,
	timestamps []time.Time,
	values []float64,
) (*models.DataLayer, error) {
	// Handle length mismatch (some tools may return fewer points)
	minLen := len(timestamps)
	if len(values) < minLen {
		minLen = len(values)
	}

	// Create TimeSeriesData from the results
	tsData, err := timeseries.CreateFromTimestampsAndValues(
		timestamps[:minLen],
		values[:minLen],
		"value",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create time series data: %w", err)
	}

	// Create DataFrame
	dataframe, err := s.dataframeService.CreateFromGotaDataFrame(projectId, layerName, tsData)
	if err != nil {
		return nil, fmt.Errorf("failed to create dataframe: %w", err)
	}

	// Create Layer
	layer, err := s.dataLayerService.Create(projectId, layerName)
	if err != nil {
		return nil, fmt.Errorf("failed to create layer: %w", err)
	}

	// Associate DataFrame with Layer
	layer.DataFrameId = &dataframe.DataFrameId
	if err := s.dataLayerService.Save(layer); err != nil {
		return nil, fmt.Errorf("failed to associate dataframe with layer: %w", err)
	}

	return layer, nil
}

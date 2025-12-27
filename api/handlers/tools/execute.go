package tools

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nathanaday/iot-data-sandbox/api/handlers"
)

// ToolExecuteRequest represents the request to execute a tool on a layer
type ToolExecuteRequest struct {
	ToolName      string                 `json:"tool_name"`
	SourceLayerId int64                  `json:"source_layer_id"`
	ProjectId     int64                  `json:"project_id"`
	OutputName    string                 `json:"output_name"`
	Parameters    map[string]interface{} `json:"parameters"`
}

// ToolExecuteResponse represents the response from executing a tool
type ToolExecuteResponse struct {
	Success       bool           `json:"success"`
	Layer         *LayerResponse `json:"layer,omitempty"`
	ResultType    string         `json:"result_type"`
	RawResult     interface{}    `json:"raw_result,omitempty"`
	ResultSummary ResultSummary  `json:"result_summary"`
}

// LayerResponse mirrors the data layer response structure
type LayerResponse struct {
	DataLayerId int64  `json:"data_layer_id"`
	ProjectId   int64  `json:"project_id"`
	DataFrameId *int64 `json:"dataframe_id,omitempty"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	ZIndex      int    `json:"z_index"`
	IsVisible   bool   `json:"is_visible"`
}

// ResultSummary contains summary information about the execution result
type ResultSummary struct {
	Rows    int    `json:"rows,omitempty"`
	Message string `json:"message,omitempty"`
}

// ExecuteTool godoc
// @Summary Execute a tool on layer data
// @Description Execute a registered tool using data from a source layer and optionally create a new layer with the results
// @Tags tools
// @Accept json
// @Produce json
// @Param request body ToolExecuteRequest true "Tool execution request"
// @Success 200 {object} ToolExecuteResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 404 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/tools/execute [post]
func (h *ToolHandler) ExecuteTool(w http.ResponseWriter, r *http.Request) {
	var request ToolExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to decode request: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if request.ToolName == "" {
		handlers.RespondError(w, "tool_name is required", http.StatusBadRequest)
		return
	}
	if request.SourceLayerId == 0 {
		handlers.RespondError(w, "source_layer_id is required", http.StatusBadRequest)
		return
	}
	if request.ProjectId == 0 {
		handlers.RespondError(w, "project_id is required", http.StatusBadRequest)
		return
	}
	if request.OutputName == "" {
		// Generate default output name
		request.OutputName = fmt.Sprintf("%s Output", request.ToolName)
	}

	// Initialize parameters if nil
	if request.Parameters == nil {
		request.Parameters = make(map[string]interface{})
	}

	// Execute the tool
	result, err := h.toolCallService.ExecuteToolOnLayer(
		request.ToolName,
		request.SourceLayerId,
		request.ProjectId,
		request.OutputName,
		request.Parameters,
	)
	if err != nil {
		handlers.RespondError(w, fmt.Sprintf("Tool execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Build response
	response := ToolExecuteResponse{
		Success:    true,
		ResultType: result.ResultType,
		RawResult:  result.RawResult,
		ResultSummary: ResultSummary{
			Rows:    result.RowCount,
			Message: result.Message,
		},
	}

	// Include layer info if a layer was created
	if result.Layer != nil {
		response.Layer = &LayerResponse{
			DataLayerId: result.Layer.DataLayerId,
			ProjectId:   result.Layer.ProjectId,
			DataFrameId: result.Layer.DataFrameId,
			Name:        result.Layer.Name,
			Color:       result.Layer.Color,
			ZIndex:      result.Layer.ZIndex,
			IsVisible:   result.Layer.IsVisible,
		}
	}

	handlers.RespondJSON(w, response, http.StatusOK)
}

package tools

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nathanaday/iot-data-sandbox/api/handlers"
)

// Swagger request/response types

type ToolCallRequest struct {
	ToolName   string                 `json:"toolName"`
	Parameters map[string]interface{} `json:"parameters"`
}

type ToolCallResponse struct {
	Result interface{} `json:"result"`
}

// CallTool godoc
// @Summary Execute a registered tool
// @Description Call a tool by name with the provided parameters
// @Tags tools
// @Accept json
// @Produce json
// @Param request body ToolCallRequest true "Tool call request"
// @Success 200 {object} ToolCallResponse
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/tools/call [post]
func (h *ToolHandler) CallTool(w http.ResponseWriter, r *http.Request) {
	var request ToolCallRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to decode request: %v", err), http.StatusBadRequest)
		return
	}

	result, err := h.toolCallService.CallTool(request.ToolName, request.Parameters)
	if err != nil {
		handlers.RespondError(w, fmt.Sprintf("Tool execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	handlers.RespondJSON(w, ToolCallResponse{Result: result}, http.StatusOK)
}

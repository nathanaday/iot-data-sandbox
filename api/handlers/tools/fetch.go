package tools

import (
	"fmt"
	"net/http"

	"github.com/nathanaday/iot-data-sandbox/api/handlers"
)

// GetToolManifest godoc
// @Summary Fetch all tools from the Tool module
// @Description Use the ToolHandler and ToolCallService to fetch all registered tools from "internal/tools" Go path
// @Tags tools
// @Produce json
// @Success 200 {object}
// @Failure 400 {object}
// @Failure 500 {object}
// @Router /api/tools [get]
func (h *ToolHandler) GetAllToolManifests(w http.ResponseWriter, r *http.Request) {

	tools, err := h.toolCallService.GetToolManifest()
	if err != nil {
		handlers.RespondError(w, fmt.Sprintf("Failed to load datasources: %v", err), http.StatusInternalServerError)
		return
	}

	handlers.RespondJSON(w, tools, http.StatusOK)
}

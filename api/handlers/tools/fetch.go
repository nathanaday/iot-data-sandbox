package tools

import (
	"net/http"

	"github.com/nathanaday/iot-data-sandbox/api/handlers"
	"github.com/nathanaday/iot-data-sandbox/internal/tools"
)

// Swagger response types

type ToolManifestListResponse struct {
	Tools []tools.ToolManifest `json:"tools"`
}

// GetAllToolManifests godoc
// @Summary Fetch all tools from the Tool module
// @Description Use the ToolHandler and ToolCallService to fetch all registered tools from "internal/tools" Go path
// @Tags tools
// @Produce json
// @Success 200 {array} tools.ToolManifest
// @Failure 400 {object} handlers.ErrorResponse
// @Failure 500 {object} handlers.ErrorResponse
// @Router /api/tools [get]
func (h *ToolHandler) GetAllToolManifests(w http.ResponseWriter, r *http.Request) {
	manifests := h.toolCallService.GetToolManifests()
	handlers.RespondJSON(w, manifests, http.StatusOK)
}

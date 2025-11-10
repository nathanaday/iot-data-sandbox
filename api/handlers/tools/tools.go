package tools

import (
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/services"
)

type ToolHandler struct {
	toolCallService *services.ToolCallService
}

func NewToolHandler(store *persistence.Store) *ToolHandler {
	return &ToolHandler{
		toolCallService: services.NewToolCallService(store),
	}
}

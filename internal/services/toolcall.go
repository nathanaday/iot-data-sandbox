package services

import (
	"encoding/json"

	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/tools"
)

type ToolCallService struct {
	store *persistence.Store
}

func NewToolCallService(store *persistence.Store) *ToolCallService {
	return &ToolCallService{
		store: store,
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

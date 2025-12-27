package tools

import "fmt"

type ToolCategory string

const (
	CategoryAnalysis  ToolCategory = "analysis"
	CategoryFilter    ToolCategory = "filter"
	CategoryTransform ToolCategory = "transform"
	CategoryAI        ToolCategory = "ai"
	CategoryOther     ToolCategory = "other"
)

type ToolManifest struct {
	Name          string                `json:"name"`
	Description   string                `json:"description"`
	Category      ToolCategory          `json:"category"`
	Documentation string                `json:"documentation"`
	Parameters    []ParameterDefinition `json:"parameters"`
	Examples      []string              `json:"examples,omitempty"`
}

type ParameterDefinition struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// ToolFunc is the standard signature all tool implementations must conform to.
// Parameters are passed as a map (typically from JSON), results returned as interface{}.
type ToolFunc func(params map[string]interface{}) (interface{}, error)

// Manifest registry
var toolRegistry = make(map[string]ToolManifest)

// Tool implementation registry
var toolImplementationRegistry = make(map[string]ToolFunc)

func RegisterTool(manifest ToolManifest, implementation ToolFunc) {
	toolRegistry[manifest.Name] = manifest
	toolImplementationRegistry[manifest.Name] = implementation
}

// Get all tool manifests for API export
func GetAllToolManifests() []ToolManifest {
	manifests := make([]ToolManifest, 0, len(toolRegistry))
	for _, manifest := range toolRegistry {
		manifests = append(manifests, manifest)
	}
	return manifests
}

func GetImplementationByName(toolName string) (ToolFunc, error) {
	implementation, ok := toolImplementationRegistry[toolName]
	if !ok {
		return nil, fmt.Errorf("implementation not found: %s", toolName)
	}
	return implementation, nil
}

// CallTool is a convenience function to look up and execute a tool
func CallTool(toolName string, params map[string]interface{}) (interface{}, error) {
	impl, err := GetImplementationByName(toolName)
	if err != nil {
		return nil, err
	}
	return impl(params)
}

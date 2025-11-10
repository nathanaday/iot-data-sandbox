package tools

type ToolManifest struct {
	Name          string                `json:"name"`
	Description   string                `json:"description"`
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

// Global registry
var toolRegistry = make(map[string]ToolManifest)

func RegisterTool(manifest ToolManifest, implementation interface{}) {
	toolRegistry[manifest.Name] = manifest
	// TODO register with LangChain here
}

// Get all tool manifests for API export
func GetAllToolManifests() []ToolManifest {
	manifests := make([]ToolManifest, 0, len(toolRegistry))
	for _, manifest := range toolRegistry {
		manifests = append(manifests, manifest)
	}
	return manifests
}

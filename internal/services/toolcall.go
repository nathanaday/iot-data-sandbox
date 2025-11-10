package services

import (
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

func (s *ToolCallService) GetToolManifest() ([]*tools.ToolManifest, error) {
	manifests := tools.GetAllToolManifests()
	tools := make([]*tools.ToolManifest, len(manifests))
	for i, manifest := range manifests {
		tools[i] = &manifest
	}
	return tools, nil
}

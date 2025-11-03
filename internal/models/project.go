package models

import "time"

type Project struct {
	ProjectId   int64
	Name        string
	WhenCreated time.Time

	// In-memory only (not persisted directly)
	Layers []*DataLayer
}

// GetLayerByName finds a layer by name in the in-memory collection
func (p *Project) GetLayerByName(name string) *DataLayer {
	for _, layer := range p.Layers {
		if layer.Name == name {
			return layer
		}
	}
	return nil
}

// GetLayerByIndex returns a layer by its index in the collection
func (p *Project) GetLayerByIndex(index int) *DataLayer {
	if index < 0 || index >= len(p.Layers) {
		return nil
	}
	return p.Layers[index]
}

// NumLayers returns the count of layers in this project
func (p *Project) NumLayers() int {
	return len(p.Layers)
}

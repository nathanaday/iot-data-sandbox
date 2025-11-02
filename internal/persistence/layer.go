package persistence

import (
	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/schemas"
)

// SaveLayer inserts or updates a DataLayer
func (s *Store) SaveLayer(layer *schemas.DataLayerSchema) error {
	if layer.DataLayerId == 0 {
		result, err := s.db.Exec(`
            INSERT INTO data_layers (project_id, data_source_id, name)
            VALUES (?, ?, ?)`,
			layer.ProjectId, layer.DataSourceId, layer.Name,
		)
		if err != nil {
			return err
		}
		layer.DataLayerId, _ = result.LastInsertId()
	} else {
		_, err := s.db.Exec(`
            UPDATE data_layers
            SET project_id=?, data_source_id=?, name=?
            WHERE data_layer_id=?`,
			layer.ProjectId, layer.DataSourceId, layer.Name, layer.DataLayerId,
		)
		return err
	}
	return nil
}

// LoadLayer retrieves a DataLayer by ID
func (s *Store) LoadLayer(id int64) (*schemas.DataLayerSchema, error) {
	layer := &schemas.DataLayerSchema{}
	err := s.db.QueryRow(`
        SELECT data_layer_id, project_id, data_source_id, name
        FROM data_layers WHERE data_layer_id=?`, id,
	).Scan(&layer.DataLayerId, &layer.ProjectId, &layer.DataSourceId, &layer.Name)

	if err != nil {
		return nil, err
	}
	return layer, nil
}

// LoadAllLayers retrieves all DataLayers
func (s *Store) LoadAllLayers() ([]*schemas.DataLayerSchema, error) {
	rows, err := s.db.Query(`
        SELECT data_layer_id, project_id, data_source_id, name
        FROM data_layers ORDER BY data_layer_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var layers []*schemas.DataLayerSchema
	for rows.Next() {
		layer := &schemas.DataLayerSchema{}
		if err := rows.Scan(&layer.DataLayerId, &layer.ProjectId, &layer.DataSourceId, &layer.Name); err != nil {
			return nil, err
		}
		layers = append(layers, layer)
	}
	return layers, rows.Err()
}

// LoadLayersByProjectId retrieves all DataLayers for a specific project
func (s *Store) LoadLayersByProjectId(projectId int64) ([]*schemas.DataLayerSchema, error) {
	rows, err := s.db.Query(`
        SELECT data_layer_id, project_id, data_source_id, name
        FROM data_layers WHERE project_id=? ORDER BY data_layer_id`, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var layers []*schemas.DataLayerSchema
	for rows.Next() {
		layer := &schemas.DataLayerSchema{}
		if err := rows.Scan(&layer.DataLayerId, &layer.ProjectId, &layer.DataSourceId, &layer.Name); err != nil {
			return nil, err
		}
		layers = append(layers, layer)
	}
	return layers, rows.Err()
}

// LoadLayersByDataSourceId retrieves all DataLayers that use a specific datasource
func (s *Store) LoadLayersByDataSourceId(dataSourceId int64) ([]*schemas.DataLayerSchema, error) {
	rows, err := s.db.Query(`
        SELECT data_layer_id, project_id, data_source_id, name
        FROM data_layers WHERE data_source_id=? ORDER BY data_layer_id`, dataSourceId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var layers []*schemas.DataLayerSchema
	for rows.Next() {
		layer := &schemas.DataLayerSchema{}
		if err := rows.Scan(&layer.DataLayerId, &layer.ProjectId, &layer.DataSourceId, &layer.Name); err != nil {
			return nil, err
		}
		layers = append(layers, layer)
	}
	return layers, rows.Err()
}

// LoadLayerWithDataSource retrieves a DataLayer with its associated DataSource in one operation
// This is useful since layers are the primary way to work with data
func (s *Store) LoadLayerWithDataSource(layerId int64) (*schemas.DataLayerSchema, *schemas.DataSourceSchema, error) {
	layer, err := s.LoadLayer(layerId)
	if err != nil {
		return nil, nil, err
	}

	dataSource, err := s.FindByID(layer.DataSourceId)
	if err != nil {
		return layer, nil, err
	}

	return layer, dataSource, nil
}

// LoadLayerWithProjectAndDataSource retrieves a DataLayer with both its Project and DataSource
// This provides complete context for working with a layer
func (s *Store) LoadLayerWithProjectAndDataSource(layerId int64) (*schemas.DataLayerSchema, *models.Project, *schemas.DataSourceSchema, error) {
	layer, err := s.LoadLayer(layerId)
	if err != nil {
		return nil, nil, nil, err
	}

	var project *models.Project
	if layer.ProjectId != 0 {
		project, err = s.LoadProject(layer.ProjectId)
		if err != nil {
			return layer, nil, nil, err
		}
	}

	dataSource, err := s.FindByID(layer.DataSourceId)
	if err != nil {
		return layer, project, nil, err
	}

	return layer, project, dataSource, nil
}

// DeleteLayer removes a DataLayer by ID
func (s *Store) DeleteLayer(id int64) error {
	_, err := s.db.Exec("DELETE FROM data_layers WHERE data_layer_id=?", id)
	return err
}

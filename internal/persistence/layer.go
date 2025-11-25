package persistence

import (
	"database/sql"

	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/schemas"
)

// SaveLayer inserts or updates a DataLayer
func (s *Store) SaveLayer(layer *models.DataLayer) error {
	if layer.DataLayerId == 0 {
		result, err := s.db.Exec(`
            INSERT INTO data_layers (project_id, dataframe_id, name, color, z_index, is_visible)
            VALUES (?, ?, ?, ?, ?, ?)`,
			layer.ProjectId, layer.DataFrameId, layer.Name, layer.Color, layer.ZIndex, layer.IsVisible,
		)
		if err != nil {
			return err
		}
		layer.DataLayerId, _ = result.LastInsertId()
	} else {
		_, err := s.db.Exec(`
            UPDATE data_layers
            SET project_id=?, dataframe_id=?, name=?, color=?, z_index=?, is_visible=?
            WHERE data_layer_id=?`,
			layer.ProjectId, layer.DataFrameId, layer.Name, layer.Color, layer.ZIndex, layer.IsVisible, layer.DataLayerId,
		)
		return err
	}
	return nil
}

// LoadLayer retrieves a DataLayer by ID
func (s *Store) LoadLayer(id int64) (*models.DataLayer, error) {
	layer := &models.DataLayer{}
	var dataframeId sql.NullInt64
	err := s.db.QueryRow(`
        SELECT data_layer_id, project_id, dataframe_id, name, color, z_index, is_visible
        FROM data_layers WHERE data_layer_id=?`, id,
	).Scan(&layer.DataLayerId, &layer.ProjectId, &dataframeId, &layer.Name, &layer.Color, &layer.ZIndex, &layer.IsVisible)

	if err != nil {
		return nil, err
	}

	if dataframeId.Valid {
		layer.DataFrameId = &dataframeId.Int64
	}

	return layer, nil
}

// LoadAllLayers retrieves all DataLayers
func (s *Store) LoadAllLayers() ([]*models.DataLayer, error) {
	rows, err := s.db.Query(`
        SELECT data_layer_id, project_id, dataframe_id, name, color, z_index, is_visible
        FROM data_layers ORDER BY data_layer_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var layers []*models.DataLayer
	for rows.Next() {
		layer := &models.DataLayer{}
		var dataframeId sql.NullInt64
		if err := rows.Scan(&layer.DataLayerId, &layer.ProjectId, &dataframeId, &layer.Name, &layer.Color, &layer.ZIndex, &layer.IsVisible); err != nil {
			return nil, err
		}
		if dataframeId.Valid {
			layer.DataFrameId = &dataframeId.Int64
		}
		layers = append(layers, layer)
	}
	return layers, rows.Err()
}

// LoadLayersByProjectId retrieves all DataLayers for a specific project ordered by z_index
func (s *Store) LoadLayersByProjectId(projectId int64) ([]*models.DataLayer, error) {
	rows, err := s.db.Query(`
        SELECT data_layer_id, project_id, dataframe_id, name, color, z_index, is_visible
        FROM data_layers WHERE project_id=? ORDER BY z_index, data_layer_id`, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var layers []*models.DataLayer
	for rows.Next() {
		layer := &models.DataLayer{}
		var dataframeId sql.NullInt64
		if err := rows.Scan(&layer.DataLayerId, &layer.ProjectId, &dataframeId, &layer.Name, &layer.Color, &layer.ZIndex, &layer.IsVisible); err != nil {
			return nil, err
		}
		if dataframeId.Valid {
			layer.DataFrameId = &dataframeId.Int64
		}
		layers = append(layers, layer)
	}
	return layers, rows.Err()
}

// LoadLayersByDataFrameId retrieves all DataLayers that use a specific dataframe
func (s *Store) LoadLayersByDataFrameId(dataframeId int64) ([]*models.DataLayer, error) {
	rows, err := s.db.Query(`
        SELECT data_layer_id, project_id, dataframe_id, name, color, z_index, is_visible
        FROM data_layers WHERE dataframe_id=? ORDER BY data_layer_id`, dataframeId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var layers []*models.DataLayer
	for rows.Next() {
		layer := &models.DataLayer{}
		var dataframeIdVal sql.NullInt64
		if err := rows.Scan(&layer.DataLayerId, &layer.ProjectId, &dataframeIdVal, &layer.Name, &layer.Color, &layer.ZIndex, &layer.IsVisible); err != nil {
			return nil, err
		}
		if dataframeIdVal.Valid {
			layer.DataFrameId = &dataframeIdVal.Int64
		}
		layers = append(layers, layer)
	}
	return layers, rows.Err()
}

// LoadLayerWithDataFrame retrieves a DataLayer with its associated DataFrame in one operation
// This is useful since layers are the primary way to work with data
func (s *Store) LoadLayerWithDataFrame(layerId int64) (*models.DataLayer, *schemas.DataFrameSchema, error) {
	layer, err := s.LoadLayer(layerId)
	if err != nil {
		return nil, nil, err
	}

	// If layer has no dataframe, return layer only
	if layer.DataFrameId == nil {
		return layer, nil, nil
	}

	dataframe, err := s.LoadDataFrame(*layer.DataFrameId)
	if err != nil {
		return layer, nil, err
	}

	return layer, dataframe, nil
}

// LoadLayerWithProjectAndDataFrame retrieves a DataLayer with both its Project and DataFrame
// This provides complete context for working with a layer
func (s *Store) LoadLayerWithProjectAndDataFrame(layerId int64) (*models.DataLayer, *models.Project, *schemas.DataFrameSchema, error) {
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

	// If layer has no dataframe, return layer and project only
	if layer.DataFrameId == nil {
		return layer, project, nil, nil
	}

	dataframe, err := s.LoadDataFrame(*layer.DataFrameId)
	if err != nil {
		return layer, project, nil, err
	}

	return layer, project, dataframe, nil
}

// DeleteLayer removes a DataLayer by ID
func (s *Store) DeleteLayer(id int64) error {
	_, err := s.db.Exec("DELETE FROM data_layers WHERE data_layer_id=?", id)
	return err
}

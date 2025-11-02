package persistence

import "github.com/nathanaday/iot-data-sandbox/internal/schemas"

// SaveDataSource inserts or updates a DataSource
func (s *Store) SaveDataSource(ds *schemas.DataSourceSchema) error {
	if ds.DataSourceId == 0 {
		result, err := s.db.Exec(`
            INSERT INTO data_sources (name, data_source_type, data_source_path, row_count, start_time, end_time, time_label, value_label, when_created)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			ds.Name, ds.DataSourceType, ds.DataSourcePath, ds.RowCount, ds.StartTime, ds.EndTime, ds.TimeLabel, ds.ValueLabel, ds.WhenCreated,
		)
		if err != nil {
			return err
		}
		ds.DataSourceId, _ = result.LastInsertId()
	} else {
		_, err := s.db.Exec(`
            UPDATE data_sources
            SET name=?, data_source_type=?, data_source_path=?, row_count=?, start_time=?, end_time=?, time_label=?, value_label=?, when_created=?
            WHERE data_source_id=?`,
			ds.Name, ds.DataSourceType, ds.DataSourcePath, ds.RowCount, ds.StartTime, ds.EndTime, ds.TimeLabel, ds.ValueLabel, ds.WhenCreated, ds.DataSourceId,
		)
		return err
	}
	return nil
}

// LoadDataSource retrieves a DataSource by ID
func (s *Store) LoadDataSource(id int64) (*schemas.DataSourceSchema, error) {
	ds := &schemas.DataSourceSchema{}
	err := s.db.QueryRow(`
        SELECT data_source_id, name, data_source_type, data_source_path, row_count, start_time, end_time, time_label, value_label, when_created
        FROM data_sources WHERE data_source_id=?`, id,
	).Scan(&ds.DataSourceId, &ds.Name, &ds.DataSourceType, &ds.DataSourcePath, &ds.RowCount, &ds.StartTime, &ds.EndTime, &ds.TimeLabel, &ds.ValueLabel, &ds.WhenCreated)

	if err != nil {
		return nil, err
	}
	return ds, nil
}

// LoadAllDataSources retrieves all DataSources ordered by creation date
func (s *Store) LoadAllDataSources() ([]*schemas.DataSourceSchema, error) {
	rows, err := s.db.Query(`
        SELECT data_source_id, name, data_source_type, data_source_path, row_count, start_time, end_time, time_label, value_label, when_created
        FROM data_sources ORDER BY when_created DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []*schemas.DataSourceSchema
	for rows.Next() {
		ds := &schemas.DataSourceSchema{}
		if err := rows.Scan(&ds.DataSourceId, &ds.Name, &ds.DataSourceType,
			&ds.DataSourcePath, &ds.RowCount, &ds.StartTime, &ds.EndTime, &ds.TimeLabel, &ds.ValueLabel, &ds.WhenCreated); err != nil {
			return nil, err
		}
		sources = append(sources, ds)
	}
	return sources, rows.Err()
}

// DeleteDataSource removes a DataSource by ID
func (s *Store) DeleteDataSource(id int64) error {
	_, err := s.db.Exec("DELETE FROM data_sources WHERE data_source_id=?", id)
	return err
}


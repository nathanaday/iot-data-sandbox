package persistence

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"github.com/nathanaday/iot-data-sandbox/internal/schemas"
)

// SaveDataFrame inserts or updates a DataFrame metadata record
func (s *Store) SaveDataFrame(df *schemas.DataFrameSchema) error {
	if df.DataFrameId == 0 {
		result, err := s.db.Exec(`
            INSERT INTO dataframes (project_id, name, description, column_definitions, row_count, start_time, end_time, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			df.ProjectId, df.Name, df.Description, df.ColumnDefinitions, df.RowCount, df.StartTime, df.EndTime, df.CreatedAt,
		)
		if err != nil {
			return err
		}
		df.DataFrameId, _ = result.LastInsertId()
	} else {
		_, err := s.db.Exec(`
            UPDATE dataframes
            SET project_id=?, name=?, description=?, column_definitions=?, row_count=?, start_time=?, end_time=?, created_at=?
            WHERE dataframe_id=?`,
			df.ProjectId, df.Name, df.Description, df.ColumnDefinitions, df.RowCount, df.StartTime, df.EndTime, df.CreatedAt, df.DataFrameId,
		)
		return err
	}
	return nil
}

// LoadDataFrame retrieves a DataFrame metadata record by ID
func (s *Store) LoadDataFrame(id int64) (*schemas.DataFrameSchema, error) {
	df := &schemas.DataFrameSchema{}
	err := s.db.QueryRow(`
        SELECT dataframe_id, project_id, name, description, column_definitions, row_count, start_time, end_time, created_at
        FROM dataframes WHERE dataframe_id=?`, id,
	).Scan(&df.DataFrameId, &df.ProjectId, &df.Name, &df.Description, &df.ColumnDefinitions, &df.RowCount, &df.StartTime, &df.EndTime, &df.CreatedAt)

	if err != nil {
		return nil, err
	}
	return df, nil
}

// LoadAllDataFrames retrieves all DataFrame metadata records
func (s *Store) LoadAllDataFrames() ([]*schemas.DataFrameSchema, error) {
	rows, err := s.db.Query(`
        SELECT dataframe_id, project_id, name, description, column_definitions, row_count, start_time, end_time, created_at
        FROM dataframes ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dataframes []*schemas.DataFrameSchema
	for rows.Next() {
		df := &schemas.DataFrameSchema{}
		if err := rows.Scan(&df.DataFrameId, &df.ProjectId, &df.Name, &df.Description, &df.ColumnDefinitions, &df.RowCount, &df.StartTime, &df.EndTime, &df.CreatedAt); err != nil {
			return nil, err
		}
		dataframes = append(dataframes, df)
	}
	return dataframes, rows.Err()
}

// LoadDataFramesByProjectId retrieves all DataFrames for a specific project
func (s *Store) LoadDataFramesByProjectId(projectId int64) ([]*schemas.DataFrameSchema, error) {
	rows, err := s.db.Query(`
        SELECT dataframe_id, project_id, name, description, column_definitions, row_count, start_time, end_time, created_at
        FROM dataframes WHERE project_id=? ORDER BY created_at DESC`, projectId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dataframes []*schemas.DataFrameSchema
	for rows.Next() {
		df := &schemas.DataFrameSchema{}
		if err := rows.Scan(&df.DataFrameId, &df.ProjectId, &df.Name, &df.Description, &df.ColumnDefinitions, &df.RowCount, &df.StartTime, &df.EndTime, &df.CreatedAt); err != nil {
			return nil, err
		}
		dataframes = append(dataframes, df)
	}
	return dataframes, rows.Err()
}

// CreateDataFrameTable creates a dynamic table for storing DataFrame data
// columnNames should include the value column names (e.g., ["value1", "value2"])
func (s *Store) CreateDataFrameTable(dataframeId int64, columnNames []string) error {
	tableName := getDataFrameTableName(dataframeId)

	// Build column definitions
	columnDefs := []string{"id INTEGER PRIMARY KEY", "timestamp DATETIME NOT NULL"}
	for _, colName := range columnNames {
		columnDefs = append(columnDefs, fmt.Sprintf("%s REAL", colName))
	}

	createTableSQL := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            %s
        )`, tableName, strings.Join(columnDefs, ", "))

	if _, err := s.db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create dataframe table: %w", err)
	}

	// Create index on timestamp for efficient querying
	indexSQL := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_timestamp ON %s(timestamp)`, tableName, tableName)
	if _, err := s.db.Exec(indexSQL); err != nil {
		return fmt.Errorf("failed to create timestamp index: %w", err)
	}

	return nil
}

// InsertDataFrameData inserts data from a Gota DataFrame into the dynamic table
// The gotaDF must have a "timestamp" column and one or more value columns
func (s *Store) InsertDataFrameData(dataframeId int64, gotaDF dataframe.DataFrame) error {
	tableName := getDataFrameTableName(dataframeId)
	columnNames := gotaDF.Names()

	// Filter out timestamp column to get value columns
	var valueColumns []string
	for _, col := range columnNames {
		if col != "timestamp" {
			valueColumns = append(valueColumns, col)
		}
	}

	if len(valueColumns) == 0 {
		return fmt.Errorf("no value columns found in dataframe")
	}

	// Build INSERT statement
	placeholders := []string{"?", "?"}
	for range valueColumns {
		placeholders = append(placeholders, "?")
	}

	insertSQL := fmt.Sprintf(`INSERT INTO %s (id, timestamp, %s) VALUES (%s)`,
		tableName,
		strings.Join(valueColumns, ", "),
		strings.Join(placeholders, ", "))

	// Prepare statement for bulk insert
	stmt, err := s.db.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// Begin transaction for bulk insert
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	txStmt := tx.Stmt(stmt)

	// Extract data from Gota DataFrame and insert
	timestampSeries := gotaDF.Col("timestamp")
	timestamps := timestampSeries.Records()

	// Get all value series
	valueSeries := make([]series.Series, len(valueColumns))
	for i, colName := range valueColumns {
		valueSeries[i] = gotaDF.Col(colName)
	}

	// Insert each row (skip header row at index 0)
	for i := 1; i < len(timestamps); i++ {
		rowID := i

		// Parse timestamp
		ts, err := time.Parse(time.RFC3339, timestamps[i])
		if err != nil {
			return fmt.Errorf("failed to parse timestamp at row %d: %w", i, err)
		}

		// Build values slice
		values := make([]interface{}, 2+len(valueColumns))
		values[0] = rowID
		values[1] = ts

		// Extract value columns
		for j, vs := range valueSeries {
			valStr := vs.Records()[i]
			val, err := strconv.ParseFloat(valStr, 64)
			if err != nil {
				return fmt.Errorf("failed to parse value at row %d, column %s: %w", i, valueColumns[j], err)
			}
			values[2+j] = val
		}

		if _, err := txStmt.Exec(values...); err != nil {
			return fmt.Errorf("failed to insert row %d: %w", i, err)
		}
	}

	return tx.Commit()
}

// LoadDataFrameData loads data from the dynamic table into a Gota DataFrame
func (s *Store) LoadDataFrameData(dataframeId int64, columnNames []string) (dataframe.DataFrame, error) {
	tableName := getDataFrameTableName(dataframeId)

	// Build SELECT statement
	selectColumns := []string{"timestamp"}
	selectColumns = append(selectColumns, columnNames...)

	selectSQL := fmt.Sprintf(`SELECT %s FROM %s ORDER BY timestamp`, strings.Join(selectColumns, ", "), tableName)

	rows, err := s.db.Query(selectSQL)
	if err != nil {
		return dataframe.DataFrame{}, fmt.Errorf("failed to query dataframe data: %w", err)
	}
	defer rows.Close()

	// Collect all data
	var timestamps []string
	valueData := make([][]string, len(columnNames))
	for i := range valueData {
		valueData[i] = []string{columnNames[i]} // Header
	}

	timestamps = append(timestamps, "timestamp") // Header

	for rows.Next() {
		var ts time.Time
		values := make([]interface{}, len(columnNames))
		valuePointers := make([]interface{}, len(columnNames)+1)
		valuePointers[0] = &ts

		for i := range values {
			valuePointers[i+1] = &values[i]
		}

		if err := rows.Scan(valuePointers...); err != nil {
			return dataframe.DataFrame{}, fmt.Errorf("failed to scan row: %w", err)
		}

		timestamps = append(timestamps, ts.Format(time.RFC3339))

		for i, val := range values {
			if floatVal, ok := val.(float64); ok {
				valueData[i] = append(valueData[i], fmt.Sprintf("%f", floatVal))
			} else {
				valueData[i] = append(valueData[i], "")
			}
		}
	}

	if err := rows.Err(); err != nil {
		return dataframe.DataFrame{}, fmt.Errorf("error iterating rows: %w", err)
	}

	// Create series
	seriesList := []series.Series{series.New(timestamps, series.String, "timestamp")}
	for i, colName := range columnNames {
		seriesList = append(seriesList, series.New(valueData[i], series.Float, colName))
	}

	return dataframe.New(seriesList...), nil
}

// UpdateDataFrameData replaces all data in the dynamic table
func (s *Store) UpdateDataFrameData(dataframeId int64, gotaDF dataframe.DataFrame) error {
	tableName := getDataFrameTableName(dataframeId)

	// Delete existing data
	deleteSQL := fmt.Sprintf(`DELETE FROM %s`, tableName)
	if _, err := s.db.Exec(deleteSQL); err != nil {
		return fmt.Errorf("failed to delete existing data: %w", err)
	}

	// Insert new data
	return s.InsertDataFrameData(dataframeId, gotaDF)
}

// DeleteDataFrame removes the DataFrame metadata and drops the dynamic table
func (s *Store) DeleteDataFrame(id int64) error {
	tableName := getDataFrameTableName(id)

	// Drop the dynamic table first
	dropSQL := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, tableName)
	if _, err := s.db.Exec(dropSQL); err != nil {
		return fmt.Errorf("failed to drop dataframe table: %w", err)
	}

	// Delete metadata
	if _, err := s.db.Exec("DELETE FROM dataframes WHERE dataframe_id=?", id); err != nil {
		return fmt.Errorf("failed to delete dataframe metadata: %w", err)
	}

	return nil
}

// getDataFrameTableName returns the dynamic table name for a DataFrame
func getDataFrameTableName(dataframeId int64) string {
	return fmt.Sprintf("timeseries_%d", dataframeId)
}

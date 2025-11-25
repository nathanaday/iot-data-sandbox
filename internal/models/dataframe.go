package models

import (
	"encoding/json"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/nathanaday/iot-data-sandbox/internal/schemas"
)

// ColumnDefinition describes a column in the DataFrame
type ColumnDefinition struct {
	Name         string `json:"name"`          // Standardized name (e.g., "value1", "value2")
	Type         string `json:"type"`          // Data type (e.g., "float64", "string")
	OriginalName string `json:"original_name"` // Original name from CSV (e.g., "temperature")
}

// DataFrame is the application-level model that represents time-series data
// Data is stored in SQLite in a dynamic table (timeseries_<dataframe_id>)
type DataFrame struct {
	// Metadata (persisted in dataframes table)
	DataFrameId       int64
	ProjectId         int64
	Name              string
	Description       string
	ColumnDefinitions []ColumnDefinition
	RowCount          int
	StartTime         *time.Time
	EndTime           *time.Time
	CreatedAt         time.Time

	// In-memory data (loaded from dynamic table when needed)
	Data dataframe.DataFrame
}

// ToSchema converts the DataFrame model to a schema for SQLite persistence
func (df *DataFrame) ToSchema() *schemas.DataFrameSchema {
	columnDefsJSON, _ := json.Marshal(df.ColumnDefinitions)

	return &schemas.DataFrameSchema{
		DataFrameId:       df.DataFrameId,
		ProjectId:         df.ProjectId,
		Name:              df.Name,
		Description:       df.Description,
		ColumnDefinitions: string(columnDefsJSON),
		RowCount:          df.RowCount,
		StartTime:         df.StartTime,
		EndTime:           df.EndTime,
		CreatedAt:         df.CreatedAt,
	}
}

// FromSchema populates the DataFrame metadata from a schema
func (df *DataFrame) FromSchema(schema *schemas.DataFrameSchema) error {
	df.DataFrameId = schema.DataFrameId
	df.ProjectId = schema.ProjectId
	df.Name = schema.Name
	df.Description = schema.Description
	df.RowCount = schema.RowCount
	df.StartTime = schema.StartTime
	df.EndTime = schema.EndTime
	df.CreatedAt = schema.CreatedAt

	if schema.ColumnDefinitions != "" {
		if err := json.Unmarshal([]byte(schema.ColumnDefinitions), &df.ColumnDefinitions); err != nil {
			return err
		}
	}

	return nil
}

// GetRowCount returns the number of rows in the DataFrame
func (df *DataFrame) GetRowCount() int {
	return df.RowCount
}

// GetTimeRange returns the start and end time of the DataFrame
func (df *DataFrame) GetTimeRange() (*time.Time, *time.Time) {
	return df.StartTime, df.EndTime
}

// GetColumnNames returns a list of all column names (excluding timestamp)
func (df *DataFrame) GetColumnNames() []string {
	names := make([]string, len(df.ColumnDefinitions))
	for i, col := range df.ColumnDefinitions {
		names[i] = col.Name
	}
	return names
}

// GetOriginalColumnNames returns a list of original column names from the source
func (df *DataFrame) GetOriginalColumnNames() []string {
	names := make([]string, len(df.ColumnDefinitions))
	for i, col := range df.ColumnDefinitions {
		names[i] = col.OriginalName
	}
	return names
}

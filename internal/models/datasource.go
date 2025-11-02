package models

import (
	"time"

	"github.com/nathanaday/iot-data-sandbox/internal/schemas"
)

// DataEntry represents a single time series data point
type DataEntry struct {
	Timestamp time.Time
	Value     float64
}

// DataSource is the application-level model that holds both metadata and in-memory data
// It is a pure domain model with no infrastructure dependencies
type DataSource struct {
	// Metadata
	DataSourceId   int64
	Name           string
	DataSourceType int
	DataSourcePath string
	TimeLabel      string
	ValueLabel     string
	WhenCreated    time.Time

	// In-memory
	Data []DataEntry
}

// ToSchema converts the DataSource model to a schema for SQLite persistence
func (ds *DataSource) ToSchema() *schemas.DataSourceSchema {
	var startTime, endTime *time.Time

	if len(ds.Data) > 0 {
		startTime = &ds.Data[0].Timestamp
		endTime = &ds.Data[len(ds.Data)-1].Timestamp
	}

	return &schemas.DataSourceSchema{
		DataSourceId:   ds.DataSourceId,
		Name:           ds.Name,
		DataSourceType: ds.DataSourceType,
		DataSourcePath: ds.DataSourcePath,
		RowCount:       len(ds.Data),
		StartTime:      startTime,
		EndTime:        endTime,
		TimeLabel:      ds.TimeLabel,
		ValueLabel:     ds.ValueLabel,
		WhenCreated:    ds.WhenCreated,
	}
}

// FromSchema populates the DataSource metadata from a schema
func (ds *DataSource) FromSchema(schema *schemas.DataSourceSchema) {
	ds.DataSourceId = schema.DataSourceId
	ds.Name = schema.Name
	ds.DataSourceType = schema.DataSourceType
	ds.DataSourcePath = schema.DataSourcePath
	ds.TimeLabel = schema.TimeLabel
	ds.ValueLabel = schema.ValueLabel
	ds.WhenCreated = schema.WhenCreated
}

func (ds *DataSource) GetRowCount() int {
	return len(ds.Data)
}

func (ds *DataSource) GetTimeRange() (start *time.Time, end *time.Time) {
	if len(ds.Data) == 0 {
		return nil, nil
	}
	return &ds.Data[0].Timestamp, &ds.Data[len(ds.Data)-1].Timestamp
}

var DataSourceTypes = map[int]string{
	0: "csv",
}

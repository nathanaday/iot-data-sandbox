package models

import (
	"time"

	"github.com/nathanaday/iot-data-sandbox/internal/schemas"
)

type DataSource struct {
	DataSourceId     int64
	Project          *Project
	Name             string
	DataSourceType   int
	DataSourcePath   string
	RowCount         int
	StartTime        *time.Time
	EndTime          *time.Time
	TimeLabel        string
	ValueLabel       string
	WhenCreated      time.Time
}

func (ds *DataSource) ToSchema() *schemas.DataSourceSchema {
	s := &schemas.DataSourceSchema{
		DataSourceId:   ds.DataSourceId,
		Name:           ds.Name,
		DataSourceType: ds.DataSourceType,
		DataSourcePath: ds.DataSourcePath,
		RowCount:       ds.RowCount,
		StartTime:      ds.StartTime,
		EndTime:        ds.EndTime,
		TimeLabel:      ds.TimeLabel,
		ValueLabel:     ds.ValueLabel,
		WhenCreated:    ds.WhenCreated,
	}

	if ds.Project != nil {
		s.ProjectId = ds.Project.ProjectId
	}

	return s
}

func (ds *DataSource) FromSchema(schema *schemas.DataSourceSchema) {
	ds.DataSourceId = schema.DataSourceId
	ds.Name = schema.Name
	ds.DataSourceType = schema.DataSourceType
	ds.DataSourcePath = schema.DataSourcePath
	ds.RowCount = schema.RowCount
	ds.StartTime = schema.StartTime
	ds.EndTime = schema.EndTime
	ds.TimeLabel = schema.TimeLabel
	ds.ValueLabel = schema.ValueLabel
	ds.WhenCreated = schema.WhenCreated
	// Note: Project object is not populated here, must be set separately if needed
	ds.Project = nil
}

var DataSourceTypes = map[int]string{
	0: "csv",
}

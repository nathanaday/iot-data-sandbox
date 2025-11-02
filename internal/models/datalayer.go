package models

import "github.com/nathanaday/iot-data-sandbox/internal/schemas"

type DataLayer struct {
	DataLayerId int64
	Name        string
	Project     *Project
	DataSource  *DataSource
}

func (dl *DataLayer) ToSchema() *schemas.DataLayerSchema {
	s := &schemas.DataLayerSchema{
		DataLayerId: dl.DataLayerId,
		Name:        dl.Name,
	}

	if dl.Project != nil {
		s.ProjectId = dl.Project.ProjectId
	}

	if dl.DataSource != nil {
		s.DataSourceId = dl.DataSource.DataSourceId
	}

	return s
}

func (dl *DataLayer) FromSchema(schema *schemas.DataLayerSchema) {
	dl.DataLayerId = schema.DataLayerId
	dl.Name = schema.Name
	dl.Project = nil    // use SetProject
	dl.DataSource = nil // use SetDataSource
}

func (dl *DataLayer) SetProject(project *Project) {
	dl.Project = project
}

func (dl *DataLayer) SetDataSource(dataSource *DataSource) {
	dl.DataSource = dataSource
}

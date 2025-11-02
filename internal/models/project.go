package models

import (
	"time"

	"github.com/nathanaday/iot-data-sandbox/internal/schemas"
)

type Project struct {
	ProjectId   int64
	Name        string
	WhenCreated time.Time
}

func (p *Project) ToSchema() *schemas.ProjectSchema {
	return &schemas.ProjectSchema{
		ProjectId:   p.ProjectId,
		Name:        p.Name,
		WhenCreated: p.WhenCreated,
	}
}

func (p *Project) FromSchema(schema *schemas.ProjectSchema) {
	p.ProjectId = schema.ProjectId
	p.Name = schema.Name
	p.WhenCreated = schema.WhenCreated
}

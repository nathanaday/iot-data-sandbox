package models

import (
	"time"

	"github.com/nathanaday/iot-data-sandbox/internal/schemas"
)

type Tool struct {
	ToolId       int64
	Name         string
	FxName       string
	TimeoutS     int
	IsEnabled    bool
	WhenLastCall *time.Time
	NumCalls     int
	MaxCalls     *int
	NumCallReset *int
	AuthProps    *ToolAuthProps
}

type ToolAuthProps struct {
	ToolId         int64
	HashedApiKey   *string
	HashedUsername *string
	HashedPassword *string
}

func (t *Tool) ToSchema() *schemas.ToolSchema {
	ts := &schemas.ToolSchema{
		ToolId:       t.ToolId,
		Name:         t.Name,
		FxName:       t.FxName,
		TimeoutS:     t.TimeoutS,
		IsEnabled:    t.IsEnabled,
		WhenLastCall: t.WhenLastCall,
		NumCalls:     t.NumCalls,
		MaxCalls:     t.MaxCalls,
		NumCallReset: t.NumCallReset,
	}

	if t.AuthProps != nil {
		ts.AuthProps = &schemas.ToolAuthPropsSchema{
			ToolId:         t.AuthProps.ToolId,
			HashedApiKey:   t.AuthProps.HashedApiKey,
			HashedUsername: t.AuthProps.HashedUsername,
			HashedPassword: t.AuthProps.HashedPassword,
		}
	}

	return ts
}

func (t *Tool) FromSchema(schema *schemas.ToolSchema) {
	t.ToolId = schema.ToolId
	t.Name = schema.Name
	t.FxName = schema.FxName
	t.TimeoutS = schema.TimeoutS
	t.IsEnabled = schema.IsEnabled
	t.WhenLastCall = schema.WhenLastCall
	t.NumCalls = schema.NumCalls
	t.MaxCalls = schema.MaxCalls
	t.NumCallReset = schema.NumCallReset

	if schema.AuthProps != nil {
		t.AuthProps = &ToolAuthProps{
			ToolId:         schema.AuthProps.ToolId,
			HashedApiKey:   schema.AuthProps.HashedApiKey,
			HashedUsername: schema.AuthProps.HashedUsername,
			HashedPassword: schema.AuthProps.HashedPassword,
		}
	} else {
		t.AuthProps = nil
	}
}

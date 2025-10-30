package models

import "time"

type ToolModel struct {
	ToolId       int64
	Name         string
	FxName       string
	TimeoutS     int
	IsEnabled    bool
	WhenLastCall *time.Time // pointer to handle NULL
	NumCalls     int
	MaxCalls     *int // pointer to handle NULL
	NumCallReset *int // pointer to handle NULL
	AuthProps    *ToolAuthProps
}

type ToolAuthProps struct {
	ToolId         int64
	HashedApiKey   *string // pointer to handle NULL
	HashedUsername *string
	HashedPassword *string
}


package models

import "time"

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


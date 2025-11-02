package schemas

import "time"

type ToolSchema struct {
	ToolId       int64
	Name         string
	FxName       string
	TimeoutS     int
	IsEnabled    bool
	WhenLastCall *time.Time
	NumCalls     int
	MaxCalls     *int
	NumCallReset *int
	AuthProps    *ToolAuthPropsSchema
}

type ToolAuthPropsSchema struct {
	ToolId         int64
	HashedApiKey   *string
	HashedUsername *string
	HashedPassword *string
}

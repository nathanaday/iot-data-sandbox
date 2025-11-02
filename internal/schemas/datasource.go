package schemas

import "time"

type DataSourceSchema struct {
	DataSourceId   int64
	ProjectId      int64
	Name           string
	DataSourceType int
	DataSourcePath string
	RowCount       int
	StartTime      *time.Time
	EndTime        *time.Time
	TimeLabel      string
	ValueLabel     string
	WhenCreated    time.Time
}

var DataSourceTypes = map[int]string{
	0: "csv",
}

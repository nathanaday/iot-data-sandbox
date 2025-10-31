package models

import "time"

type DataSource struct {
	DataSourceId   int64
	Name           string
	DataSourceType int
	DataSourcePath string
	WhenCreated    time.Time
}

var DataSourceTypes = map[int]string{
	0: "csv",
}


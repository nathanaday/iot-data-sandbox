package models

import "time"

type Project struct {
	ProjectId   int64
	Name        string
	WhenCreated time.Time
}

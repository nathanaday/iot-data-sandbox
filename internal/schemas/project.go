package schemas

import "time"

type ProjectSchema struct {
	ProjectId   int64
	Name        string
	WhenCreated time.Time
}

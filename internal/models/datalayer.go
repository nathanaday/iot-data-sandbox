package models

import (
	"time"

	"github.com/go-gota/gota/dataframe"
)

type DataLayer struct {
	DataLayerId int64
	ProjectId   int64
	DataFrameId *int64
	Name        string

	// UI/UX properties
	Color     string
	ZIndex    int
	IsVisible bool

	// In-memory relationships (not persisted)
	Project   *Project
	DataFrame *DataFrame
}

// GetData returns the Gota DataFrame from the underlying DataFrame
func (dl *DataLayer) GetData() dataframe.DataFrame {
	if dl.DataFrame == nil {
		return dataframe.DataFrame{}
	}
	return dl.DataFrame.Data
}

// GetTimeRange returns the actual time range of the underlying data
func (dl *DataLayer) GetTimeRange() (start *time.Time, end *time.Time) {
	if dl.DataFrame == nil {
		return nil, nil
	}
	return dl.DataFrame.GetTimeRange()
}

// IsHidden returns true if the layer is hidden in the UI
func (dl *DataLayer) IsHidden() bool {
	return !dl.IsVisible
}

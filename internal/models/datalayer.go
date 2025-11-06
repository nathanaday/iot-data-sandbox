package models

import "time"

type DataLayer struct {
	DataLayerId  int64
	ProjectId    int64
	DataSourceId *int64
	Name         string

	// UI/UX properties
	Color            string
	ZIndex           int
	IsVisible        bool
	DisplayStartTime *time.Time
	DisplayEndTime   *time.Time

	// In-memory relationships (not persisted)
	Project    *Project
	DataSource *DataSource
}

// GetData returns the time series data points from the underlying DataSource
func (dl *DataLayer) GetData() []DataEntry {
	if dl.DataSource == nil {
		return []DataEntry{}
	}
	return dl.DataSource.Data
}

// GetTimeRange returns the actual time range of the underlying data
func (dl *DataLayer) GetTimeRange() (start *time.Time, end *time.Time) {
	if dl.DataSource == nil {
		return nil, nil
	}
	return dl.DataSource.GetTimeRange()
}

// GetDisplayTimeRange returns the display window (defaults to actual time range if not set)
func (dl *DataLayer) GetDisplayTimeRange() (start *time.Time, end *time.Time) {
	if dl.DisplayStartTime != nil && dl.DisplayEndTime != nil {
		return dl.DisplayStartTime, dl.DisplayEndTime
	}
	return dl.GetTimeRange()
}

// IsHidden returns true if the layer is hidden in the UI
func (dl *DataLayer) IsHidden() bool {
	return !dl.IsVisible
}

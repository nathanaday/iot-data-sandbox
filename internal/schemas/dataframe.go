package schemas

import "time"

// DataFrameSchema represents the DataFrame metadata stored in the dataframes table
// The actual time-series data is stored in a separate dynamic table (timeseries_<dataframe_id>)
type DataFrameSchema struct {
	DataFrameId       int64      `json:"dataframe_id"`
	ProjectId         int64      `json:"project_id"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	ColumnDefinitions string     `json:"column_definitions"` // JSON string of []ColumnDefinition
	RowCount          int        `json:"row_count"`
	StartTime         *time.Time `json:"start_time"`
	EndTime           *time.Time `json:"end_time"`
	CreatedAt         time.Time  `json:"created_at"`
}

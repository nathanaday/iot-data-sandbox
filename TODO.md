# TODO

### New API calls

### Get layer data metadata

**Path**: `GET {{baseUrl}}/api/layers/:id/data/metadata`
**Description**: Get metadata details about a specific layer's data source
**Response**

(JSON representation of DataSourceMetadata struct)
```go
	DataSourceId int64      `json:"data_source_id"`
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	RowCount     int        `json:"row_count"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	TimeLabel    string     `json:"time_label"`
	ValueLabel   string     `json:"value_label"`
	WhenCreated  time.Time  `json:"when_created"`
```

### Preview csv data

**Path**: `POST {{baseUrl}}/api/ui/preview_data`
**Description**: Preview a CSV file before it's saved to a project
**Body**: (Multipart form for .csv file upload)
**Response**

(JSON representation of minimal DataSourceMetadata struct)
```go
	Type         string     `json:"type"`
	RowCount     int        `json:"row_count"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	TimeLabel    string     `json:"time_label"`
	ValueLabel   string     `json:"value_label"`
```



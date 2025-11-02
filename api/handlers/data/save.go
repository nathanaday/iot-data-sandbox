package data

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response types for API endpoints

type UploadResponse struct {
	DataSourceId int64      `json:"data_source_id"`
	Name         string     `json:"name"`
	RowCount     int        `json:"row_count"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	TimeLabel    string     `json:"time_label"`
	ValueLabel   string     `json:"value_label"`
	WhenCreated  time.Time  `json:"when_created"`
}

type DataSourceListResponse struct {
	DataSources []DataSourceMetadata `json:"data_sources"`
}

type DataSourceMetadata struct {
	DataSourceId int64      `json:"data_source_id"`
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	RowCount     int        `json:"row_count"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	TimeLabel    string     `json:"time_label"`
	ValueLabel   string     `json:"value_label"`
	WhenCreated  time.Time  `json:"when_created"`
}

type DataQueryResponse struct {
	Data      []DataPoint `json:"data"`
	RowCount  int         `json:"row_count"`
	StartTime time.Time   `json:"start_time"`
	EndTime   time.Time   `json:"end_time"`
}

type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Helper functions

func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
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

// Project response types

type ProjectResponse struct {
	ProjectId   int64     `json:"project_id"`
	Name        string    `json:"name"`
	WhenCreated time.Time `json:"when_created"`
	LayerCount  int       `json:"layer_count"`
}

type ProjectListResponse struct {
	Projects []ProjectResponse `json:"projects"`
}

type CreateProjectRequest struct {
	Name string `json:"name"`
}

// Layer response types

type LayerResponse struct {
	DataLayerId  int64  `json:"data_layer_id"`
	ProjectId    int64  `json:"project_id"`
	DataSourceId *int64 `json:"data_source_id,omitempty"`
	Name         string `json:"name"`
	Color        string `json:"color"`
	ZIndex       int    `json:"z_index"`
	IsVisible    bool   `json:"is_visible"`
}

type LayerListResponse struct {
	Layers []LayerResponse `json:"layers"`
}

type CreateLayerRequest struct {
	Name string `json:"name"`
}

type LoadCSVRequest struct {
	CSVFilename string `json:"csv_filename"`
}

type UpdateColorRequest struct {
	Color string `json:"color"`
}

type UpdateVisibilityRequest struct {
	IsVisible bool `json:"is_visible"`
}

type DuplicateLayerRequest struct {
	NewName string `json:"new_name"`
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
package models

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/schemas"
	"github.com/nathanaday/iot-data-sandbox/internal/storage"
	"github.com/nathanaday/iot-data-sandbox/internal/timeseries"
)

// DataEntry represents a single time series data point
type DataEntry struct {
	Timestamp time.Time
	Value     float64
}

// DataSource is the application-level model that holds both metadata and in-memory data
type DataSource struct {
	// Metadata
	DataSourceId   int64
	Name           string
	DataSourceType int
	DataSourcePath string
	TimeLabel      string
	ValueLabel     string
	WhenCreated    time.Time

	// In-memory
	Data []DataEntry

	// (injected)
	store     *persistence.Store
	fileStore *storage.FileStore
}

// FromCSV creates a new DataSource by loading data from a CSV file
// It automatically creates the SQLite metadata record and loads data into memory
// csvFilename should be just the filename (e.g., "data_123.csv"), not the full path
func FromCSV(name string, csvFilename string, store *persistence.Store, fileStore *storage.FileStore) (*DataSource, error) {
	// Get full path and load/validate CSV using timeseries package
	fullPath := fileStore.GetFilePath(csvFilename)
	tsData, err := timeseries.LoadAndValidateCSV(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load CSV: %w", err)
	}

	// Parse the already-validated data from the DataFrame
	data, err := parseDataFrameToEntries(tsData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV data: %w", err)
	}

	ds := &DataSource{
		Name:           name,
		DataSourceType: 0, // CSV
		DataSourcePath: csvFilename, // Store only filename, not full path
		TimeLabel:      tsData.TimeLabel,
		ValueLabel:     tsData.ValueLabel,
		WhenCreated:    time.Now(),
		Data:           data,
		store:          store,
		fileStore:      fileStore,
	}

	schema := ds.ToSchema()
	if err := store.SaveDataSource(schema); err != nil {
		return nil, fmt.Errorf("failed to save datasource metadata: %w", err)
	}
	ds.DataSourceId = schema.DataSourceId

	return ds, nil
}

// LoadFromStorage loads an existing DataSource from storage by ID
func LoadFromStorage(id int64, store *persistence.Store, fileStore *storage.FileStore) (*DataSource, error) {
	schema, err := store.LoadDataSource(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load datasource metadata: %w", err)
	}

	ds := &DataSource{
		store:     store,
		fileStore: fileStore,
	}
	ds.FromSchema(schema)

	filePath := fileStore.GetFilePath(ds.DataSourcePath)
	if err := ds.loadDataFromCSV(filePath); err != nil {
		return nil, fmt.Errorf("failed to load data from CSV: %w", err)
	}

	return ds, nil
}

// Save persists both the metadata to SQLite and the data to the CSV file
func (ds *DataSource) Save() error {
	schema := ds.ToSchema()
	if err := ds.store.SaveDataSource(schema); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	filePath := ds.fileStore.GetFilePath(ds.DataSourcePath)
	if err := ds.writeDataToCSV(filePath); err != nil {
		return fmt.Errorf("failed to write CSV: %w", err)
	}

	return nil
}

// loadDataFromCSV reads the CSV file and populates the Data slice
// Uses the timeseries package for validation and parsing
func (ds *DataSource) loadDataFromCSV(filePath string) error {
	// Use timeseries package for validation and parsing
	tsData, err := timeseries.LoadAndValidateCSV(filePath)
	if err != nil {
		return err
	}

	// Parse the validated DataFrame into DataEntry slice
	data, err := parseDataFrameToEntries(tsData)
	if err != nil {
		return err
	}

	ds.Data = data
	return nil
}

// parseDataFrameToEntries converts timeseries DataFrame to []DataEntry
// This centralizes CSV parsing logic to avoid duplication
func parseDataFrameToEntries(tsData *timeseries.TimeSeriesData) ([]DataEntry, error) {
	data := make([]DataEntry, 0, tsData.RowCount)
	timestampRecords := tsData.DataFrame.Col(tsData.TimeLabel).Records()
	valueRecords := tsData.DataFrame.Col(tsData.ValueLabel).Records()

	// Skip header row (index 0)
	for i := 1; i < len(timestampRecords); i++ {
		ts, err := time.Parse(time.RFC3339, timestampRecords[i])
		if err != nil {
			continue // Skip invalid timestamps
		}

		val, err := strconv.ParseFloat(valueRecords[i], 64)
		if err != nil {
			continue // Skip invalid values
		}

		data = append(data, DataEntry{
			Timestamp: ts,
			Value:     val,
		})
	}

	return data, nil
}

// writeDataToCSV writes the in-memory data back to the CSV file
func (ds *DataSource) writeDataToCSV(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{ds.TimeLabel, ds.ValueLabel}); err != nil {
		return err
	}

	for _, entry := range ds.Data {
		record := []string{
			entry.Timestamp.Format(time.RFC3339),
			strconv.FormatFloat(entry.Value, 'f', -1, 64),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// ToSchema converts the DataSource model to a schema for SQLite persistence
func (ds *DataSource) ToSchema() *schemas.DataSourceSchema {
	var startTime, endTime *time.Time

	if len(ds.Data) > 0 {
		startTime = &ds.Data[0].Timestamp
		endTime = &ds.Data[len(ds.Data)-1].Timestamp
	}

	return &schemas.DataSourceSchema{
		DataSourceId:   ds.DataSourceId,
		Name:           ds.Name,
		DataSourceType: ds.DataSourceType,
		DataSourcePath: ds.DataSourcePath,
		RowCount:       len(ds.Data),
		StartTime:      startTime,
		EndTime:        endTime,
		TimeLabel:      ds.TimeLabel,
		ValueLabel:     ds.ValueLabel,
		WhenCreated:    ds.WhenCreated,
	}
}

// FromSchema populates the DataSource metadata from a schema
func (ds *DataSource) FromSchema(schema *schemas.DataSourceSchema) {
	ds.DataSourceId = schema.DataSourceId
	ds.Name = schema.Name
	ds.DataSourceType = schema.DataSourceType
	ds.DataSourcePath = schema.DataSourcePath
	ds.TimeLabel = schema.TimeLabel
	ds.ValueLabel = schema.ValueLabel
	ds.WhenCreated = schema.WhenCreated
}

func (ds *DataSource) GetRowCount() int {
	return len(ds.Data)
}

func (ds *DataSource) GetTimeRange() (start *time.Time, end *time.Time) {
	if len(ds.Data) == 0 {
		return nil, nil
	}
	return &ds.Data[0].Timestamp, &ds.Data[len(ds.Data)-1].Timestamp
}

var DataSourceTypes = map[int]string{
	0: "csv",
}

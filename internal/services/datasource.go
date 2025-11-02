package services

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/storage"
	"github.com/nathanaday/iot-data-sandbox/internal/timeseries"
)

// DataSourceService provides business operations for DataSource entities
type DataSourceService struct {
	store     *persistence.Store
	fileStore *storage.FileStore
}

// NewDataSourceService creates a new DataSourceService
func NewDataSourceService(store *persistence.Store, fileStore *storage.FileStore) *DataSourceService {
	return &DataSourceService{
		store:     store,
		fileStore: fileStore,
	}
}

// CreateFromCSV creates a new DataSource by loading data from a CSV file
// It automatically creates the SQLite metadata record and loads data into memory
// csvFilename should be just the filename (e.g., "data_123.csv"), not the full path
func (s *DataSourceService) CreateFromCSV(name string, csvFilename string) (*models.DataSource, error) {
	// Get full path and load/validate CSV using timeseries package
	fullPath := s.fileStore.GetFilePath(csvFilename)
	tsData, err := timeseries.LoadAndValidateCSV(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load CSV: %w", err)
	}

	// Parse the already-validated data from the DataFrame
	data, err := parseDataFrameToEntries(tsData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV data: %w", err)
	}

	ds := &models.DataSource{
		Name:           name,
		DataSourceType: 0, // CSV
		DataSourcePath: csvFilename, // Store only filename, not full path
		TimeLabel:      tsData.TimeLabel,
		ValueLabel:     tsData.ValueLabel,
		WhenCreated:    time.Now(),
		Data:           data,
	}

	schema := ds.ToSchema()
	if err := s.store.Save(schema); err != nil {
		return nil, fmt.Errorf("failed to save datasource metadata: %w", err)
	}
	ds.DataSourceId = schema.DataSourceId

	return ds, nil
}

// LoadByID loads an existing DataSource from storage by ID
// It retrieves metadata from SQLite and loads the actual data from the CSV file
func (s *DataSourceService) LoadByID(id int64) (*models.DataSource, error) {
	schema, err := s.store.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load datasource metadata: %w", err)
	}

	ds := &models.DataSource{}
	ds.FromSchema(schema)

	filePath := s.fileStore.GetFilePath(ds.DataSourcePath)
	if err := loadDataFromCSV(ds, filePath); err != nil {
		return nil, fmt.Errorf("failed to load data from CSV: %w", err)
	}

	return ds, nil
}

// Save persists both the metadata to SQLite and the data to the CSV file
func (s *DataSourceService) Save(ds *models.DataSource) error {
	schema := ds.ToSchema()
	if err := s.store.Save(schema); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	filePath := s.fileStore.GetFilePath(ds.DataSourcePath)
	if err := writeDataToCSV(ds, filePath); err != nil {
		return fmt.Errorf("failed to write CSV: %w", err)
	}

	return nil
}

// loadDataFromCSV reads the CSV file and populates the Data slice
// Uses the timeseries package for validation and parsing
func loadDataFromCSV(ds *models.DataSource, filePath string) error {
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
func parseDataFrameToEntries(tsData *timeseries.TimeSeriesData) ([]models.DataEntry, error) {
	data := make([]models.DataEntry, 0, tsData.RowCount)
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

		data = append(data, models.DataEntry{
			Timestamp: ts,
			Value:     val,
		})
	}

	return data, nil
}

// writeDataToCSV writes the in-memory data back to the CSV file
func writeDataToCSV(ds *models.DataSource, filePath string) error {
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

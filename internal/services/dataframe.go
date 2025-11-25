package services

import (
	"fmt"
	"io"
	"time"

	"github.com/go-gota/gota/dataframe"
	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/timeseries"
)

type DataFrameService struct {
	store *persistence.Store
}

func NewDataFrameService(store *persistence.Store) *DataFrameService {
	return &DataFrameService{
		store: store,
	}
}

// CreateFromCSV creates a new DataFrame by parsing CSV data
// The CSV is not saved to filesystem; data is directly inserted into SQLite
func (s *DataFrameService) CreateFromCSV(projectId int64, name string, csvReader io.Reader) (*models.DataFrame, error) {
	// Load and validate CSV using timeseries package
	tsData, err := timeseries.LoadAndValidateCSVFromReader(csvReader)
	if err != nil {
		return nil, fmt.Errorf("failed to load CSV: %w", err)
	}

	return s.CreateFromGotaDataFrame(projectId, name, tsData)
}

// CreateFromGotaDataFrame creates a new DataFrame from a timeseries.TimeSeriesData
func (s *DataFrameService) CreateFromGotaDataFrame(projectId int64, name string, tsData *timeseries.TimeSeriesData) (*models.DataFrame, error) {
	gotaDF := tsData.DataFrame

	// Extract column definitions (skip timestamp column)
	columnNames := gotaDF.Names()
	var columnDefs []models.ColumnDefinition
	var valueColumns []string

	for _, colName := range columnNames {
		if colName == "timestamp" {
			continue
		}
		columnDefs = append(columnDefs, models.ColumnDefinition{
			Name:         colName,
			Type:         "float64",
			OriginalName: colName, // Could map from tsData if we tracked original names
		})
		valueColumns = append(valueColumns, colName)
	}

	// Create DataFrame model
	df := &models.DataFrame{
		ProjectId:         projectId,
		Name:              name,
		Description:       "",
		ColumnDefinitions: columnDefs,
		RowCount:          tsData.RowCount - 1, // Subtract header row
		StartTime:         &tsData.StartTime,
		EndTime:           &tsData.EndTime,
		CreatedAt:         time.Now(),
	}

	// Save metadata
	schema := df.ToSchema()
	if err := s.store.SaveDataFrame(schema); err != nil {
		return nil, fmt.Errorf("failed to save dataframe metadata: %w", err)
	}
	df.DataFrameId = schema.DataFrameId

	// Create dynamic table
	if err := s.store.CreateDataFrameTable(df.DataFrameId, valueColumns); err != nil {
		// Rollback: delete metadata
		s.store.DeleteDataFrame(df.DataFrameId)
		return nil, fmt.Errorf("failed to create dataframe table: %w", err)
	}

	// Insert data
	if err := s.store.InsertDataFrameData(df.DataFrameId, gotaDF); err != nil {
		// Rollback: delete everything
		s.store.DeleteDataFrame(df.DataFrameId)
		return nil, fmt.Errorf("failed to insert dataframe data: %w", err)
	}

	// Load the data back into the model
	df.Data = gotaDF

	return df, nil
}

// LoadByID loads a DataFrame including its data
func (s *DataFrameService) LoadByID(id int64) (*models.DataFrame, error) {
	// Load metadata
	schema, err := s.store.LoadDataFrame(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load dataframe metadata: %w", err)
	}

	df := &models.DataFrame{}
	if err := df.FromSchema(schema); err != nil {
		return nil, fmt.Errorf("failed to parse dataframe schema: %w", err)
	}

	// Extract column names from definitions
	columnNames := df.GetColumnNames()

	// Load data from dynamic table
	gotaDF, err := s.store.LoadDataFrameData(id, columnNames)
	if err != nil {
		return nil, fmt.Errorf("failed to load dataframe data: %w", err)
	}

	df.Data = gotaDF

	return df, nil
}

// GetMetadata loads only the DataFrame metadata without loading the data
func (s *DataFrameService) GetMetadata(id int64) (*models.DataFrame, error) {
	schema, err := s.store.LoadDataFrame(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load dataframe metadata: %w", err)
	}

	df := &models.DataFrame{}
	if err := df.FromSchema(schema); err != nil {
		return nil, fmt.Errorf("failed to parse dataframe schema: %w", err)
	}

	return df, nil
}

// Update replaces the data in an existing DataFrame
func (s *DataFrameService) Update(id int64, gotaDF dataframe.DataFrame) error {
	// Load existing metadata to validate
	df, err := s.GetMetadata(id)
	if err != nil {
		return fmt.Errorf("failed to load dataframe: %w", err)
	}

	// Extract new metadata from the updated DataFrame
	columnNames := gotaDF.Names()
	var valueColumns []string
	for _, colName := range columnNames {
		if colName != "timestamp" {
			valueColumns = append(valueColumns, colName)
		}
	}

	// Update column definitions if structure changed
	var newColumnDefs []models.ColumnDefinition
	for _, colName := range valueColumns {
		newColumnDefs = append(newColumnDefs, models.ColumnDefinition{
			Name:         colName,
			Type:         "float64",
			OriginalName: colName,
		})
	}

	// Update row count and time range
	df.ColumnDefinitions = newColumnDefs
	df.RowCount = gotaDF.Nrow() - 1 // Subtract header row

	// Extract time range from Gota DataFrame
	startTime, endTime, err := extractTimeRange(gotaDF)
	if err == nil {
		df.StartTime = &startTime
		df.EndTime = &endTime
	}

	// Save updated metadata
	schema := df.ToSchema()
	if err := s.store.SaveDataFrame(schema); err != nil {
		return fmt.Errorf("failed to update dataframe metadata: %w", err)
	}

	// Update data in the dynamic table
	if err := s.store.UpdateDataFrameData(id, gotaDF); err != nil {
		return fmt.Errorf("failed to update dataframe data: %w", err)
	}

	return nil
}

// Delete removes a DataFrame and its data
func (s *DataFrameService) Delete(id int64) error {
	return s.store.DeleteDataFrame(id)
}

// ListByProject returns all DataFrames for a project (metadata only)
func (s *DataFrameService) ListByProject(projectId int64) ([]*models.DataFrame, error) {
	schemas, err := s.store.LoadDataFramesByProjectId(projectId)
	if err != nil {
		return nil, fmt.Errorf("failed to load dataframes: %w", err)
	}

	dataframes := make([]*models.DataFrame, 0, len(schemas))
	for _, schema := range schemas {
		df := &models.DataFrame{}
		if err := df.FromSchema(schema); err != nil {
			continue
		}
		dataframes = append(dataframes, df)
	}

	return dataframes, nil
}

// extractTimeRange extracts start and end time from a Gota DataFrame
func extractTimeRange(df dataframe.DataFrame) (time.Time, time.Time, error) {
	timestampSeries := df.Col("timestamp")
	records := timestampSeries.Records()

	if len(records) <= 1 {
		return time.Time{}, time.Time{}, fmt.Errorf("insufficient data")
	}

	var minTime, maxTime time.Time

	for i := 1; i < len(records); i++ {
		t, err := time.Parse(time.RFC3339, records[i])
		if err != nil {
			continue
		}

		if i == 1 || t.Before(minTime) {
			minTime = t
		}
		if i == 1 || t.After(maxTime) {
			maxTime = t
		}
	}

	return minTime, maxTime, nil
}

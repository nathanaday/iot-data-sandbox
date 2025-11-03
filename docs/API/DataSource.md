# DataSource API

## Service: `DataSourceService`

### Methods

- **`CreateFromCSV(name, csvFilename)`** - Creates a new DataSource from a CSV file. Automatically validates CSV, loads data into memory, and saves metadata to SQLite. Returns the created DataSource.

- **`LoadByID(id)`** - Loads an existing DataSource by ID. Retrieves metadata from SQLite and loads CSV data into memory. Returns the DataSource with in-memory data.

- **`Save(dataSource)`** - Persists both metadata (to SQLite) and data (to CSV file). Use after making mutations to the dataset.

## Model: `DataSource`

### Fields

- **`DataSourceId int64`** - Unique identifier
- **`Name string`** - Human-readable name
- **`DataSourceType int`** - Type of data source (0 = CSV)
- **`DataSourcePath string`** - CSV filename (not full path)
- **`TimeLabel string`** - Column name for timestamps (e.g., "time")
- **`ValueLabel string`** - Column name for values (e.g., "value")
- **`WhenCreated time.Time`** - Creation timestamp
- **`Data []DataEntry`** - In-memory slice of time series data points

### Methods

- **`GetRowCount()`** - Returns the current number of data entries
- **`GetTimeRange()`** - Returns start and end timestamps (or nil if empty)
- **`ToSchema()`** - Converts to schema for persistence (internal use)
- **`FromSchema(schema)`** - Populates from schema (internal use)

## DataEntry Struct

- **`Timestamp time.Time`** - Data point timestamp
- **`Value float64`** - Data point numeric value

## Example Usage

```go
// Create service
dsService := services.NewDataSourceService(store, fileStore)

// Create datasource from CSV
ds, err := dsService.CreateFromCSV("My Data", "data_123.csv")

// Load existing datasource with data
ds, err := dsService.LoadByID(1)

// Access data
for _, entry := range ds.Data {
    fmt.Printf("%v: %f\n", entry.Timestamp, entry.Value)
}

// Get metadata
count := ds.GetRowCount()
start, end := ds.GetTimeRange()
```

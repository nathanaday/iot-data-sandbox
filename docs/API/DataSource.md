# DataSource Model API

  ## Constructor Functions

  - **`FromCSV(name, csvFilename, store, fileStore)`** - Creates a new DataSource from a CSV file. Automatically validates CSV, loads data into memory, and creates SQLite
  metadata record.

  - **`LoadFromStorage(id, store, fileStore)`** - Loads an existing DataSource by ID. Retrieves metadata from SQLite and loads CSV data into memory.

  ## Instance Methods

  - **`Save()`** - Persists both metadata (to SQLite) and data (to CSV file). Use after making mutations to the dataset.

  - **`ToSchema()`** - Converts DataSource to schema format for SQLite persistence. Automatically calculates row count and time range from in-memory data.

  - **`FromSchema(schema)`** - Populates DataSource metadata from a schema. Note: Only loads metadata, not actual data.

  - **`GetRowCount()`** - Returns the current number of data entries in memory.

  - **`GetTimeRange()`** - Returns the start and end timestamps of the dataset (or nil if empty).

  ## Fields

  - **`Data []DataEntry`** - In-memory slice of time series data points
  - **`DataSourceId`** - Unique identifier
  - **`Name`** - Human-readable name
  - **`DataSourcePath`** - CSV filename (not full path)
  - **`TimeLabel`** - Column name for timestamps
  - **`ValueLabel`** - Column name for values
  - **`WhenCreated`** - Creation timestamp

  ## DataEntry Struct

  - **`Timestamp time.Time`** - Data point timestamp
  - **`Value float64`** - Data point value

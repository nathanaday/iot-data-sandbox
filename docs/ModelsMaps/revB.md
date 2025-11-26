# IoT Data Sandbox - Architecture Documentation (Rev B)

---

## Table of Contents
1. [Architecture Overview](#architecture-overview)
2. [Core Data Structures](#core-data-structures)
3. [Database Schema](#database-schema)
4. [Service Layer](#service-layer)
5. [Data Flow Examples](#data-flow-examples)

---

## Architecture Overview

### Design Philosophy

```
CSV Upload → Parse & Validate → Insert to SQLite → Done
                                      ↓
                            (No CSV files saved)
```

### Key Architectural Decisions

1. **Database-as-Source-of-Truth**
   - All data lives in SQLite
   - No CSV files persisted to disk
   - CSV is only used during upload (streamed directly to DB)

2. **Dynamic Table-per-DataFrame**
   - Each DataFrame gets its own SQLite table: `timeseries_<dataframe_id>`
   - Allows flexible schema (different DataFrames can have different columns)
   - Enables efficient querying and indexing per dataset

3. **Multi-Column Support**
   - DataFrames support multiple value columns (e.g., temperature, humidity, pressure)
   - All columns stored in single table with timestamp

4. **In-Place Mutations**
   - Data transformations update the DataFrame table directly
   - No CSV sync issues

---

## Core Data Structures

### 1. Project
**Purpose:** Top-level container for organizing related data layers.

**Location:** `internal/models/project.go`

```go
type Project struct {
    ProjectId   int64           // Primary key
    Name        string          // User-defined name
    WhenCreated time.Time       // Creation timestamp
    Layers      []*DataLayer    // In-memory collection (loaded on demand)
}
```

**Why:** Projects provide logical grouping for dashboards, experiments, or data analysis sessions.

---

### 2. DataFrame
**Purpose:** Represents a time-series dataset with metadata and actual data.

**Location:** `internal/models/dataframe.go`

```go
type DataFrame struct {
    // Metadata (persisted in 'dataframes' table)
    DataFrameId       int64
    ProjectId         int64                // FK to projects
    Name              string               // User-defined name
    Description       string               // Optional description
    ColumnDefinitions []ColumnDefinition   // Column metadata (JSON)
    RowCount          int                  // Number of data rows
    StartTime         *time.Time           // First timestamp
    EndTime           *time.Time           // Last timestamp
    CreatedAt         time.Time            // When created

    // In-memory (loaded from dynamic table when needed)
    Data              dataframe.DataFrame  // Gota DataFrame
}

type ColumnDefinition struct {
    Name         string  // Standardized name (e.g., "value1", "value2")
    Type         string  // Data type (e.g., "float64")
    OriginalName string  // Original CSV column name (e.g., "temperature")
}
```

**Why:**
- **Metadata separation:** DataFrame metadata is cheap to load, data is loaded only when needed
- **Multi-column tracking:** ColumnDefinitions map original CSV names to standardized DB columns
- **Flexible schema:** Each DataFrame can have different columns

---

### 3. DataLayer
**Purpose:** Visual representation of a DataFrame with display properties.

**Location:** `internal/models/datalayer.go`

```go
type DataLayer struct {
    DataLayerId int64           // Primary key
    ProjectId   int64           // FK to projects
    DataFrameId *int64          // FK to dataframes (nullable)
    Name        string          // Layer name

    // UI/UX properties
    Color       string          // Hex color (#3b82f6)
    ZIndex      int             // Stacking order
    IsVisible   bool            // Visibility toggle

    // In-memory relationships
    Project     *Project        // Back-reference
    DataFrame   *DataFrame      // Associated data
}
```

**Why:**
- **Separation of concerns:** DataFrame = data, DataLayer = presentation
- **Multiple views:** Different layers can reference the same DataFrame with different display settings
- **Nullable DataFrame:** Layers can exist without data (prepared for future assignment)

---

## Database Schema

### Tables Overview

```
projects
    ↓ (1:N)
dataframes ←─┐
    ↓        │
data_layers ─┘ (N:1)

timeseries_1  ← Dynamic table for dataframe_id=1
timeseries_2  ← Dynamic table for dataframe_id=2
...
```

### Table: `projects`
```sql
CREATE TABLE projects (
    project_id   INTEGER PRIMARY KEY AUTOINCREMENT,
    name         TEXT NOT NULL,
    when_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Table: `dataframes`
```sql
CREATE TABLE dataframes (
    dataframe_id       INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id         INTEGER NOT NULL,
    name               TEXT NOT NULL,
    description        TEXT,
    column_definitions TEXT,              -- JSON: [{name, type, originalName}, ...]
    row_count          INTEGER DEFAULT 0,
    start_time         TIMESTAMP,
    end_time           TIMESTAMP,
    created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(project_id) ON DELETE CASCADE
);
```

**Why column_definitions is JSON:**
- Flexible schema: different DataFrames have different columns
- Preserves original CSV column names for user reference
- Compact storage for metadata

### Table: `data_layers`
```sql
CREATE TABLE data_layers (
    data_layer_id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id    INTEGER,
    dataframe_id  INTEGER,
    name          TEXT NOT NULL,
    color         TEXT NOT NULL DEFAULT '#3b82f6',
    z_index       INTEGER NOT NULL DEFAULT 0,
    is_visible    BOOLEAN NOT NULL DEFAULT 1,
    FOREIGN KEY (project_id)   REFERENCES projects(project_id)     ON DELETE CASCADE,
    FOREIGN KEY (dataframe_id) REFERENCES dataframes(dataframe_id) ON DELETE CASCADE
);
```

### Dynamic Tables: `timeseries_<dataframe_id>`
**Example for dataframe_id = 1:**
```sql
CREATE TABLE timeseries_1 (
    id        INTEGER PRIMARY KEY,
    timestamp DATETIME NOT NULL,
    value1    REAL,
    value2    REAL,
    ...       -- Additional columns as needed
);
CREATE INDEX idx_timeseries_1_timestamp ON timeseries_1(timestamp);
```

**Why dynamic tables:**
- **Flexible schema:** Each dataset can have different columns
- **Performance:** Indexes per table, no giant shared table
- **Simplicity:** Standard SQL queries, no complex JSON extraction

---

## Service Layer

### Architecture Pattern
```
Controller (HTTP Handler)
    ↓
Service Layer (Business Logic)
    ↓
Persistence Layer (Store)
    ↓
SQLite Database
```

### Key Services

#### 1. DataFrameService
**Location:** `internal/services/dataframe.go`

**Responsibilities:**
- Create DataFrames from CSV uploads
- Load DataFrame data from SQLite
- Update DataFrame data in-place
- Delete DataFrames (drops table + metadata)

**Key Methods:**
```go
// Create from CSV stream (no filesystem)
CreateFromCSV(projectId int64, name string, csvReader io.Reader) (*DataFrame, error)

// Load with all data (metadata + timeseries)
LoadByID(id int64) (*DataFrame, error)

// Load metadata only (efficient for listings)
GetMetadata(id int64) (*DataFrame, error)

// Update data in-place
Update(id int64, gotaDataFrame dataframe.DataFrame) error

// Delete everything (table + metadata)
Delete(id int64) error
```

**Why this design:**
- **Streaming CSV:** `io.Reader` means no temp files
- **Lazy loading:** `GetMetadata()` avoids loading large datasets
- **In-place updates:** Enables data transformations without new tables

---

#### 2. DataLayerService
**Location:** `internal/services/datalayer.go`

**Responsibilities:**
- Create and manage layers
- Associate DataFrames with layers
- Update display properties (color, visibility, z-index)
- Load layers with their DataFrame data

**Key Methods:**
```go
// Create empty layer
Create(projectId int64, name string) (*DataLayer, error)

// Load layer with DataFrame data
LoadWithDataFrame(id int64) (*DataLayer, error)

// Load CSV directly into layer (creates DataFrame)
LoadFromCSV(layerId int64, csvReader io.Reader) error

// Update display properties
UpdateColor(layerId int64, color string) error
UpdateZIndex(layerId int64, zIndex int) error
SetVisibility(layerId int64, visible bool) error

// Duplicate layer (shares same DataFrame)
Duplicate(layerId int64, newName string) (*DataLayer, error)
```

**Why duplicate shares DataFrame:**
- Same data, different view (e.g., zoomed vs full range)
- Saves storage space
- Changes to DataFrame affect all layers

---

#### 3. ProjectService
**Location:** `internal/services/project.go`

**Responsibilities:**
- Create and manage projects
- Add layers to projects
- Load projects with all layers
- Save project state

**Key Methods:**
```go
Create(name string) (*Project, error)
LoadByID(id int64) (*Project, error)
LoadWithLayers(id int64) (*Project, error)
AddLayer(projectId int64, name string) (*DataLayer, error)
SaveAll(project *Project) error
Delete(id int64) error
```

---

## Data Flow Examples

### Example 1: CSV Upload → Visualization

```
1. User uploads CSV via API
   POST /api/dataframes
   Body: multipart/form-data with file + project_id

2. DataFrameHandler receives request
   ↓
3. DataFrameService.CreateFromCSV(projectId, name, csvReader)
   ↓
4. timeseries.LoadAndValidateCSVFromReader(csvReader)
   - Parse CSV in-memory
   - Validate structure (must have timestamp + value columns)
   - Normalize timestamps to RFC3339
   ↓
5. Persistence Layer
   a. Insert metadata → dataframes table
   b. Create dynamic table → timeseries_<id>
   c. Bulk insert data → timeseries_<id>
   ↓
6. Return DataFrame metadata to user
   {dataframe_id, name, row_count, start_time, end_time, ...}
```

**Key Points:**
- CSV never touches filesystem
- All I/O is streaming (memory-efficient)
- Transaction ensures atomicity (metadata + data created together)

---

### Example 2: Create Project with Layer

```
1. Create Project
   projectSvc.Create("My Dashboard")
   → INSERT INTO projects (name) VALUES (...)
   → Returns Project{ProjectId: 1, Name: "My Dashboard"}

2. Add Layer to Project
   projectSvc.AddLayer(projectId=1, "Temperature Sensor")
   → INSERT INTO data_layers (project_id, name, color, z_index, is_visible)
      VALUES (1, "Temperature Sensor", "#3b82f6", 0, 1)
   → Returns DataLayer{DataLayerId: 1, ProjectId: 1, DataFrameId: nil, ...}

3. Load CSV into Layer
   dataLayerSvc.LoadFromCSV(layerId=1, csvReader)
   ↓
   a. Load layer metadata
   b. Create DataFrame from CSV (see Example 1)
   c. Update layer: SET dataframe_id = <new_id>
   ↓
   Layer now has data!
```

---

### Example 3: Query Layer Data with Time Range

```
1. User requests layer data
   GET /api/layers/1/data?start_time=2024-01-01T00:00:00Z&end_time=2024-01-02T00:00:00Z

2. LayerHandler.GetLayerData()
   ↓
3. dataLayerSvc.LoadWithDataFrame(layerId=1)
   ↓
4. Persistence Layer
   a. Load layer metadata → data_layers table
   b. Load DataFrame metadata → dataframes table
   c. Load data from → timeseries_<dataframe_id> table
      SELECT timestamp, value1, value2, ...
      FROM timeseries_<id>
      ORDER BY timestamp
   ↓
5. Service returns DataLayer with populated DataFrame
   ↓
6. Handler filters by time range (in-memory)
   - Iterate through Gota DataFrame
   - Filter records by start_time/end_time
   ↓
7. Return JSON
   {
     data: [{timestamp, value}, ...],
     row_count: 150,
     start_time: "2024-01-01T00:00:00Z",
     end_time: "2024-01-02T00:00:00Z"
   }
```

**Future Optimization:**
Move time filtering to SQL query for better performance on large datasets.

---

### Example 4: Duplicate Layer (Sharing DataFrame)

```
1. User duplicates layer
   POST /api/layers/1/duplicate
   Body: {new_name: "Temperature (Zoomed)"}

2. dataLayerSvc.Duplicate(layerId=1, "Temperature (Zoomed)")
   ↓
3. Load original layer
   → DataLayer{DataLayerId: 1, DataFrameId: 5, ...}
   ↓
4. Create new layer with same dataframe_id
   INSERT INTO data_layers (project_id, dataframe_id, name, color, z_index, is_visible)
   VALUES (1, 5, "Temperature (Zoomed)", "#3b82f6", 1, 1)
   ↓
5. Return new layer
   → DataLayer{DataLayerId: 2, DataFrameId: 5, ...}

Result: Two layers, one DataFrame
- Layer 1: "Temperature Sensor" (full range)
- Layer 2: "Temperature (Zoomed)" (different zoom, same data)
- Both point to DataFrame 5
```

**Why this is useful:**
- User can have multiple visualizations of same data
- Saves storage (no data duplication)
- Changes to DataFrame affect both layers

---

## Relationship Diagram

```
┌─────────────┐
│   Project   │
│ project_id  │
│ name        │
└──────┬──────┘
       │ 1:N
       ├─────────────────────────┐
       │                         │
       ▼                         ▼
┌─────────────┐          ┌─────────────┐
│  DataFrame  │          │  DataLayer  │
│dataframe_id │◄─────────│data_layer_id│
│ project_id  │   N:1    │ project_id  │
│ name        │          │dataframe_id │
│ row_count   │          │ name        │
│ start_time  │          │ color       │
│ end_time    │          │ z_index     │
│ ...         │          │ is_visible  │
└──────┬──────┘          └─────────────┘
       │
       │ 1:1
       ▼
┌──────────────────┐
│ timeseries_<id>  │  ← Dynamic table
│ id               │
│ timestamp        │
│ value1           │
│ value2           │
│ ...              │
└──────────────────┘
```

### Relationship Rules
1. **Project → DataFrame**: One-to-Many with CASCADE delete
2. **Project → DataLayer**: One-to-Many with CASCADE delete
3. **DataFrame → DataLayer**: One-to-Many with CASCADE delete
4. **Multiple layers can share the same DataFrame**
5. **Deleting a DataFrame drops its timeseries table**

---

## Design Decisions FAQ

### Q: Why not store everything in one big table?
**A:** Scalability and flexibility. Separate tables allow:
- Different schemas per dataset (some have 2 columns, others 10)
- Independent indexing strategies
- Easier to drop/recreate specific datasets
- Better query performance (smaller indexes)

### Q: Why use Gota DataFrame in memory?
**A:** Gota provides:
- Convenient data manipulation (filter, transform)
- Type safety
- Integration with Go's type system
- Future support for complex operations (join, group, aggregate)

### Q: Why allow multiple layers per DataFrame?
**A:** User experience:
- Same data, different views (e.g., zoomed vs overview)
- Different color schemes for comparison
- Show/hide specific datasets without data duplication

### Q: Why not use JSON columns in SQLite?
**A:** Performance and simplicity:
- Standard SQL is faster than JSON extraction
- Better index support
- Simpler queries
- JSON only used for column metadata (small, infrequently accessed)

---

## File Structure Reference

```
internal/
├── models/              # Domain models (pure Go structs)
│   ├── dataframe.go     # DataFrame + ColumnDefinition
│   ├── datalayer.go     # DataLayer
│   └── project.go       # Project
│
├── schemas/             # Database schemas (DB representation)
│   └── dataframe.go     # DataFrameSchema
│
├── persistence/         # Database operations (Store)
│   ├── store.go         # Database setup + table creation
│   ├── dataframe.go     # DataFrame CRUD + dynamic table ops
│   ├── layer.go         # DataLayer CRUD
│   └── project.go       # Project CRUD
│
├── services/            # Business logic layer
│   ├── dataframe.go     # DataFrameService
│   ├── datalayer.go     # DataLayerService
│   └── project.go       # ProjectService
│
└── timeseries/          # CSV parsing and validation
    └── csv_loader.go    # LoadAndValidateCSVFromReader()

api/handlers/data/       # HTTP handlers
├── dataframe.go         # DataFrame API endpoints
├── layer_handler.go     # Layer API endpoints
├── project_handler.go   # Project API endpoints
└── ui_handler.go        # UI helper endpoints (preview)
```

---

## Summary

**Core Concept:** All time-series data lives in SQLite. DataFrames own the data, DataLayers provide visualization/presentation, Projects organize everything.

**Key Benefits:**
1. No filesystem dependencies
2. Data integrity via database transactions
3. Flexible schema per dataset
4. Efficient querying with indexes
5. Multiple views of same data (layers)

**For New Developers:**
- Start reading: `models/` → understand data structures
- Then: `services/` → understand business logic
- Finally: `persistence/` → understand database operations
- API handlers are thin wrappers around services

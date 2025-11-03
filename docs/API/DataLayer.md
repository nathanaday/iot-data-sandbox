# DataLayer API

## Service: `DataLayerService`

### Methods

- **`Create(projectId, name)`** - Creates a new DataLayer within a project with default UI properties (blue color, z-index 0, visible).

- **`LoadByID(id)`** - Loads a DataLayer by ID (metadata and UI properties only, DataSource not loaded).

- **`LoadWithDataSource(id)`** - Loads a DataLayer with its associated DataSource fully loaded (includes in-memory time series data).

- **`LoadFromCSV(layerId, csvFilename)`** - Loads CSV data into a layer by creating a new DataSource and associating it with the layer.

- **`Save(layer)`** - Persists a DataLayer (metadata and UI properties).

- **`Delete(id)`** - Removes a DataLayer by ID.

- **`UpdateDisplayWindow(layerId, start, end)`** - Updates the display time window for UI viewport control.

- **`SetVisibility(layerId, visible)`** - Shows or hides a layer in the UI.

- **`Duplicate(layerId, newName)`** - Creates a copy of a layer with a new name. Shares the same underlying DataSource.

- **`UpdateColor(layerId, color)`** - Updates the layer's color (HEX format, e.g., "#3b82f6").

- **`UpdateZIndex(layerId, zIndex)`** - Updates the layer's z-index (stacking order).

## Model: `DataLayer`

### Fields

**Core:**
- **`DataLayerId int64`** - Unique identifier
- **`ProjectId int64`** - Parent project ID
- **`DataSourceId int64`** - Associated datasource ID
- **`Name string`** - Layer name

**UI Properties:**
- **`Color string`** - HEX color (default: "#3b82f6")
- **`ZIndex int`** - Stacking order (0 = bottom, higher = top)
- **`IsVisible bool`** - Visibility flag (default: true)
- **`DisplayStartTime *time.Time`** - UI viewport start (nil = use data start)
- **`DisplayEndTime *time.Time`** - UI viewport end (nil = use data end)

**In-Memory Relationships (not persisted):**
- **`Project *Project`** - Reference to parent project
- **`DataSource *DataSource`** - Reference to underlying datasource

### Methods

- **`GetData()`** - Returns the time series data from the underlying DataSource. Returns empty slice if no datasource.

- **`GetTimeRange()`** - Returns the actual time range of the underlying data (start, end timestamps).

- **`GetDisplayTimeRange()`** - Returns the display window. Defaults to actual time range if display window not set.

- **`IsHidden()`** - Returns true if the layer is hidden in the UI.

## Example Usage

```go
// Create service
layerSvc := services.NewDataLayerService(store, dataSourceSvc)

// Create layer in project
layer, err := layerSvc.Create(projectId, "Temperature Sensor")

// Load CSV data into layer
err := layerSvc.LoadFromCSV(layer.DataLayerId, "temp_data.csv")

// Load layer with data
layer, err := layerSvc.LoadWithDataSource(layerId)

// Access data
data := layer.GetData()
for _, entry := range data {
    fmt.Printf("%v: %f\n", entry.Timestamp, entry.Value)
}

// Configure UI properties
err := layerSvc.UpdateColor(layer.DataLayerId, "#ff5733")
err := layerSvc.UpdateZIndex(layer.DataLayerId, 5)
err := layerSvc.SetVisibility(layer.DataLayerId, false) // hide

// Set display window for zooming
start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
end := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
err := layerSvc.UpdateDisplayWindow(layer.DataLayerId, &start, &end)

// Duplicate layer
layer2, err := layerSvc.Duplicate(layer.DataLayerId, "Temperature Sensor (Copy)")

// Access time ranges
actualStart, actualEnd := layer.GetTimeRange()
displayStart, displayEnd := layer.GetDisplayTimeRange()

// Check visibility
if layer.IsHidden() {
    fmt.Println("Layer is hidden")
}
```

## Workflow Example: Complete Project Setup

```go
// 1. Create project
p, _ := projectSvc.Create("My IoT Dashboard")

// 2. Add layer
layer, _ := projectSvc.AddLayer(p.ProjectId, "Sensor 1")

// 3. Load data into layer
layerSvc.LoadFromCSV(layer.DataLayerId, "sensor_data.csv")

// 4. Configure display
layerSvc.UpdateColor(layer.DataLayerId, "#4caf50")
layerSvc.SetVisibility(layer.DataLayerId, true)

// 5. Add another layer with different view
layer2, _ := layerSvc.Duplicate(layer.DataLayerId, "Sensor 1 (Zoomed)")
start := time.Now().Add(-24 * time.Hour)
end := time.Now()
layerSvc.UpdateDisplayWindow(layer2.DataLayerId, &start, &end)

// 6. Save everything
projectSvc.SaveAll(p)
```

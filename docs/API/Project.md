# Project API

## Service: `ProjectService`

### Methods

- **`Create(name)`** - Creates a new empty project. Returns the created Project.

- **`LoadByID(id)`** - Loads a Project by ID (metadata only, layers not loaded).

- **`LoadWithLayers(id)`** - Loads a Project with all its layers (ordered by z-index).

- **`LoadAll()`** - Retrieves all projects (metadata only).

- **`Save(project)`** - Persists project metadata only (does not save layers).

- **`SaveAll(project)`** - Persists project and all its layers in one operation.

- **`Delete(id)`** - Removes a project by ID. Cascade deletes all associated layers.

- **`AddLayer(projectId, layerName)`** - Creates a new layer within the project. Returns the created DataLayer.

- **`ReorderLayer(projectId, layerId, newZIndex)`** - Changes the z-index of a layer to reorder it. Automatically adjusts other layers.

- **`GetLayerCount(projectId)`** - Returns the number of layers in the project.

## Model: `Project`

### Fields

- **`ProjectId int64`** - Unique identifier
- **`Name string`** - Project name
- **`WhenCreated time.Time`** - Creation timestamp
- **`Layers []*DataLayer`** - In-memory collection of layers (not persisted directly)

### Methods

- **`GetLayerByName(name)`** - Finds a layer by name in the in-memory collection. Returns nil if not found.

- **`GetLayerByIndex(index)`** - Returns a layer by its index in the collection. Returns nil if out of bounds.

- **`NumLayers()`** - Returns the count of layers in this project.

## Example Usage

```go
// Create service
projectSvc := services.NewProjectService(store, dataLayerSvc)

// Create new project
p, err := projectSvc.Create("My IoT Project")

// Add layers
layer1, err := projectSvc.AddLayer(p.ProjectId, "Sensor Data")
layer2, err := projectSvc.AddLayer(p.ProjectId, "Processed Data")

// Load project with all layers
p, err := projectSvc.LoadWithLayers(projectId)

// Access layers
fmt.Printf("Project has %d layers\n", p.NumLayers())
layer := p.GetLayerByName("Sensor Data")
layer := p.GetLayerByIndex(0)

// Reorder layers (change stacking)
err := projectSvc.ReorderLayer(p.ProjectId, layer2.DataLayerId, 0) // move to bottom

// Save all
err := projectSvc.SaveAll(p) // saves project + all layers

// Delete project (cascades to layers)
err := projectSvc.Delete(p.ProjectId)
```

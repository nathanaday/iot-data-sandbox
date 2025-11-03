# iot-data-sandbox

This is a Go-based IoT data sandbox project that enables LLM-driven analysis and manipulation of time-series data. 

Users interact with the system through natural language prompts:
- "show me any anomalies that occurred in the last 3 days"
- "plot the moving average with a window of 6 hours"

An agentic AI agent orchestrates a registered set of tools to fulfill these requests, ranging from simple statistical analysis to AI/ML forecasting models and anomaly detection.

This project is in very early stages and not yet functional.

### References

https://github.com/tmc/langchaingo


### Generate API Docs:

```
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go
```


# Running locally

```
go run cmd/server/main.go
```

### Workflow Concept

The entity hierarchy follows the pattern shown below:
- `project` -> `datalayer` -> `datasource`

`Project` - this is the logical container for a set of "datalayers" so the use can easily use "save all" and "load all" style requests on their current workspace. In the UI, all displayed timeseries data and layers will be based on the current loaded project. Other than gluing together datalayer entities, the project does not have to do anything highly specialized

`Datalayer` - this is a container for a datasource set. It wraps the datasource and mirrors all the same features that the datasource has (save/load from csv, manage time scale, manage datapoints), but it adds some user experience and visual mix-ins: specify color, theme, and z-index, useful for displaying many layers on the same UI canvas. Also, a safe interface to allow simple convenience operations like duplicating layers, collapsing layers

`Datasource` - this entity manages save/load from csv data on disk and tracks the datasource record in sqlite. This app model contains the actual timeseries data and other metadata like length, label names, and time span.


Intended API Usage (pseudocode examples)
- Let's walk through a user starting up a fresh, new project

```go

p = Project.New("name-of-this-project") // New empty project created with name
p.numLayers() // 0
p.save() // saves all entities to storage (for a project with no layers, this is just the sqlite entry of the project itself)

layer1 = p.addLayer("name-of-new-layer")  // Layer constructor, automatically becomes a member of the project
layer1.LoadFromCSV("path_to_csv.csv") // Loads csv data into the layer via 'Datasource' entity
layer1.setColor("#454545")  // supports color HEX only for simplicity
layer1.rename("layer-1")

Datapoints[] dp = layer.getData()  // returns list of (timestamp, value) tuples corresponding to data (UI will use this to display data)


// Working with the time window
// The layer has an actual time span that it can support based on the underlying DataSource loaded from csv (example, the csv may cover time from Jan-01-2010 to Feb-01-2010)
// This is the real time frame and we always need to be able to access it for UI/display convenience
layer1.getStartTime()  // ISO timestamp (or similar)
layer1.getEndTime()

// But for actual UI display, we often change the time scale to view the data more closely; this has no effect on the stored start and end time
layer1.getDisplayStartTime(*time.Time t) // default is getStartTime
layer1.getDisplayEndTime(*time.Time t)  // default is getEndTime
layer1.setDisplayStartTime(*time.Time t) // must be GTE getStartTime
layer1.setDisplayEndTime(*time.Time t)  // must be LTE getEndTime

// Showing and hiding a layer (UI feature)
layer1.makeVisible()
layer1.makeHidden()


// If you happen to lose reference to the layer returned from the constructor, you can also access it via project
del layer1
layer1 = p.getLayerByName("layer-1")
layer1 = p.getLayerByIndex(0)

// Managing multiple layers
layer2 = p.addLayer("layer-2")
layer3 = p.addLayer("layer-3")

layers[] layers = p.getAllLayers  // Returns a list in index order
p.setLayerIndex("layer-2", 0)  // re-arranges other layers as needed; note, cannot be assigned an index out of bounds

p.save() // Save all layers, data, and info to disk

```

Future steps (do not implement):
- Manipulating time series data using analysis tools





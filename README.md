# iot-data-sandbox

<img width="1669" height="1111" alt="Screenshot 2026-01-20 at 10 47 29 AM" src="https://github.com/user-attachments/assets/74afac69-f587-4a7b-9b0e-ad786b7b91af" />

---

This is a Go-based IoT data sandbox project that enables LLM-driven analysis and manipulation of time-series data. 

Users interact with the system through natural language prompts:
- "show me any anomalies that occurred in the last 3 days"
- "plot the moving average with a window of 6 hours"

An agentic AI agent orchestrates a registered set of tools to fulfill these requests, ranging from simple statistical analysis to AI/ML forecasting models and anomaly detection.


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





NA CO2 Emissions
https://github.com/rishabh89007/Time_Series_Datasets/blob/main/NA%20Emissions.csv


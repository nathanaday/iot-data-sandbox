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




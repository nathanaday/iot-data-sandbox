# iot-data-sandbox

<img width="1669" height="1111" alt="Screenshot 2026-01-20 at 10 47 29 AM" src="https://github.com/user-attachments/assets/74afac69-f587-4a7b-9b0e-ad786b7b91af" />

---

A Go-based IoT data sandbox that enables LLM-driven analysis and manipulation of time-series data.

Users interact with the system through natural language prompts:
- "show me any anomalies that occurred in the last 3 days"
- "plot the moving average with a window of 6 hours"

An agentic AI agent orchestrates a registered set of tools to fulfill these requests, ranging from simple statistical analysis to AI/ML forecasting models and anomaly detection.

## Project Structure

| Repository | Description |
|------------|-------------|
| [iot-data-sandbox](https://github.com/nathanaday/iot-data-sandbox) | Go backend API server |
| [iot-data-sandbox-ui](https://github.com/nathanaday/iot-data-sandbox-ui) | Vue.js frontend application |

## Quick Start

### 1. Start the backend

```bash
git clone https://github.com/nathanaday/iot-data-sandbox.git
cd iot-data-sandbox
go run cmd/server/main.go
```

The server runs on `localhost:8080`.

### 2. Start the frontend

```bash
git clone https://github.com/nathanaday/iot-data-sandbox-ui.git
cd iot-data-sandbox-ui
npm install && npm run dev
```

The frontend expects the backend to be running at `localhost:8080`.

## Entity Model

The entity hierarchy follows: `project` -> `datalayer` -> `datasource`

- **Project** - Logical container for datalayers; enables save/load of entire workspaces
- **Datalayer** - Wraps a datasource with visual properties (color, theme, z-index) for UI display
- **Datasource** - Manages CSV data and tracks timeseries metadata (length, labels, time span)

## Generate API Docs

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/server/main.go
```

## References

- [langchaingo](https://github.com/tmc/langchaingo)
- [NA CO2 Emissions Dataset](https://github.com/rishabh89007/Time_Series_Datasets/blob/main/NA%20Emissions.csv)

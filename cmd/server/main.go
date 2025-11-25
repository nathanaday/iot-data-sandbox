package main

import (
	"log"

	"github.com/nathanaday/iot-data-sandbox/api"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"

	_ "github.com/nathanaday/iot-data-sandbox/docs"
)

// @title IoT Data Sandbox API
// @version 1.0
// @description API for managing and querying time series data from IoT sensors
// @description
// @description This API allows you to upload CSV files containing time series data,
// @description organize data in projects and layers, and query the data with time range filters.
// @description All data is stored in SQLite with no external file dependencies.
// @description
// @description Supported timestamp formats: ISO8601, RFC3339, Unix timestamps (seconds/milliseconds), Julian Day
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @license.name MIT
// @host localhost:8080
// @BasePath /

func main() {
	store, err := persistence.NewStore("./sandbox.db")
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}
	defer store.Close()

	log.Println("Database initialized successfully")

	router := api.SetupRouter(store)
	err = api.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

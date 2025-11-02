package main

import (
	"log"

	"github.com/nathanaday/iot-data-sandbox/api"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/storage"

	_ "github.com/nathanaday/iot-data-sandbox/docs"
)

// @title IoT Data Sandbox API
// @version 1.0
// @description API for managing and querying time series data from IoT sensors
// @description
// @description This API allows you to upload CSV files containing time series data,
// @description query the data with time range filters, and manage datasources.
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

	fileStore, err := storage.NewFileStore()
	if err != nil {
		log.Fatalf("Failed to initialize file store: %v", err)
	}
	log.Printf("File storage initialized at: %s", fileStore.GetBaseDir())

	router := api.SetupRouter(store, fileStore)
	err = api.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

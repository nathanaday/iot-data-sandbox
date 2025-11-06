package main

import (
	"log"
	"time"

	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/services"
	"github.com/nathanaday/iot-data-sandbox/internal/storage"
)

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

	dataSourceSvc := services.NewDataSourceService(store, fileStore)
	dataLayerSvc := services.NewDataLayerService(store, dataSourceSvc)
	projectSvc := services.NewProjectService(store, dataLayerSvc)

	// Project Creation
	p, err := projectSvc.Create("My IoT Dashboard")
	if err != nil {
		log.Fatalf("Failed to create project: %v", err)
	}
	log.Printf("Created project: %s (ID: %d)", p.Name, p.ProjectId)

	// Layer Creation
	layer, err := projectSvc.AddLayer(p.ProjectId, "Sensor 1")
	if err != nil {
		log.Fatalf("Failed to add layer: %v", err)
	}
	log.Printf("Created layer: %s (ID: %d)", layer.Name, layer.DataLayerId)

	// Data Loading
	err = dataLayerSvc.LoadFromCSV(layer.DataLayerId, "test_data.csv")
	if err != nil {
		log.Fatalf("Failed to load CSV: %v", err)
	}
	log.Printf("Loaded CSV data into layer")

	// Layer Display Configuration
	err = dataLayerSvc.UpdateColor(layer.DataLayerId, "#4caf50")
	if err != nil {
		log.Fatalf("Failed to update color: %v", err)
	}
	err = dataLayerSvc.SetVisibility(layer.DataLayerId, true)
	if err != nil {
		log.Fatalf("Failed to set visibility: %v", err)
	}
	log.Printf("Updated layer display properties")

	// Layer Duplication
	layer2, err := dataLayerSvc.Duplicate(layer.DataLayerId, "Sensor 1 (Zoomed)")
	if err != nil {
		log.Fatalf("Failed to duplicate layer: %v", err)
	}
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	err = dataLayerSvc.UpdateDisplayWindow(layer2.DataLayerId, &start, &end)
	if err != nil {
		log.Fatalf("Failed to update display window: %v", err)
	}
	log.Printf("Created duplicate layer with custom time window")

	// Project Saving
	err = projectSvc.SaveAll(p)
	if err != nil {
		log.Fatalf("Failed to save project: %v", err)
	}
	log.Printf("Successfully saved project with all layers")
}

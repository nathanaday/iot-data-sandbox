package main

import (
	"log"
	"os"

	"github.com/nathanaday/iot-data-sandbox/internal/jobs"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/services"
)

func main() {
	// Initialize database
	store, err := persistence.NewStore("./sandbox.db")
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}
	defer store.Close()

	// Initialize job manager
	jobManager := jobs.NewJobManager()

	// Initialize services
	dataframeSvc := services.NewDataFrameService(store)
	dataLayerSvc := services.NewDataLayerService(store, dataframeSvc)
	projectSvc := services.NewProjectService(store, dataLayerSvc, jobManager)

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

	// Data Loading - CSV is parsed and stored directly in SQLite
	// No filesystem storage - data goes directly to database
	csvFile, err := os.Open("app-files/demo-sources/test_data.csv")
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer csvFile.Close()

	err = dataLayerSvc.LoadFromCSV(layer.DataLayerId, csvFile)
	if err != nil {
		log.Fatalf("Failed to load CSV: %v", err)
	}
	log.Printf("Loaded CSV data into layer (stored in SQLite)")

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

	// Layer Duplication (shares the same DataFrame)
	layer2, err := dataLayerSvc.Duplicate(layer.DataLayerId, "Sensor 1 (Zoomed)")
	if err != nil {
		log.Fatalf("Failed to duplicate layer: %v", err)
	}
	log.Printf("Created duplicate layer: %s (ID: %d) - shares DataFrame with original", layer2.Name, layer2.DataLayerId)

	// Project Saving
	err = projectSvc.SaveAll(p)
	if err != nil {
		log.Fatalf("Failed to save project: %v", err)
	}
	log.Printf("Successfully saved project with all layers")

	// Demonstrate DataFrame independence
	log.Printf("\n--- DataFrame Architecture Demo ---")
	log.Printf("Layer 1 and Layer 2 both reference the same DataFrame")
	log.Printf("All data is stored in SQLite - no CSV files saved")
	log.Printf("DataFrame can be updated in-place, affecting all layers")
}

package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nathanaday/iot-data-sandbox/api"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/storage"
	httpSwagger "github.com/swaggo/http-swagger"

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
	store, err := persistence.NewStore("./iot-data.db")
	if err != nil {
		log.Fatalf("Failed to initialize store: %v", err)
	}
	defer store.Close()

	fileStore, err := storage.NewFileStore()
	if err != nil {
		log.Fatalf("Failed to initialize file store: %v", err)
	}

	log.Printf("File storage initialized at: %s", fileStore.GetBaseDir())

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(corsMiddleware)

	dataSourceHandler := api.NewDataSourceHandler(store, fileStore)

	r.Route("/api/datasources", func(r chi.Router) {
		r.Post("/", dataSourceHandler.UploadCSV)
		r.Get("/", dataSourceHandler.ListDataSources)
		r.Get("/{id}", dataSourceHandler.GetDataSource)
		r.Get("/{id}/data", dataSourceHandler.QueryData)
		r.Delete("/{id}", dataSourceHandler.DeleteDataSource)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	addr := ":8080"
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
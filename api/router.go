package api

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nathanaday/iot-data-sandbox/api/handlers/data"
	"github.com/nathanaday/iot-data-sandbox/api/handlers/tools"
	"github.com/nathanaday/iot-data-sandbox/internal/jobs"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	httpSwagger "github.com/swaggo/http-swagger"
)

func ListenAndServe(addr string, r *chi.Mux) error {

	if addr == "" {
		addr = ":8080" // Default
	}
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		return err
	}
	return nil
}

func SetupRouter(store *persistence.Store) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(corsMiddleware)

	// Create job manager for async operations
	jobManager := jobs.NewJobManager()

	// DataFrame routes (replaces datasources)
	dataframeHandler := data.NewDataFrameHandler(store)
	r.Route("/api/dataframes", func(r chi.Router) {
		r.Post("/", dataframeHandler.UploadCSV)
		r.Get("/", dataframeHandler.ListDataFrames)
		r.Get("/{id}", dataframeHandler.GetDataFrame)
		r.Delete("/{id}", dataframeHandler.DeleteDataFrame)
	})

	projectHandler := data.NewProjectHandler(store, jobManager)
	r.Route("/api/projects", func(r chi.Router) {
		r.Post("/", projectHandler.CreateProject)
		r.Get("/", projectHandler.ListProjects)
		r.Get("/{id}", projectHandler.GetProject)
		r.Delete("/{id}", projectHandler.DeleteProject)
		r.Post("/{id}/layers", projectHandler.AddLayer)
		r.Get("/{id}/layers", projectHandler.GetProjectLayers)
		r.Post("/{id}/load-csv", projectHandler.LoadCSV)
		r.Get("/{id}/load-csv/status", projectHandler.GetLoadCSVStatus)
	})

	layerHandler := data.NewLayerHandler(store)
	r.Route("/api/layers", func(r chi.Router) {
		r.Get("/{id}", layerHandler.GetLayer)
		r.Delete("/{id}", layerHandler.DeleteLayer)
		r.Post("/{id}/load-csv", layerHandler.LoadCSV)
		r.Put("/{id}/color", layerHandler.UpdateColor)
		r.Put("/{id}/visibility", layerHandler.UpdateVisibility)
		r.Post("/{id}/duplicate", layerHandler.DuplicateLayer)
		r.Get("/{id}/data/metadata", layerHandler.GetLayerDataMetadata)
		r.Get("/{id}/data", layerHandler.GetLayerData)
	})

	uiHandler := data.NewUIHandler(store)
	r.Route("/api/ui", func(r chi.Router) {
		r.Post("/preview_data", uiHandler.PreviewCSV)
	})

	toolHandler := tools.NewToolHandler(store)
	r.Route("/api/tools", func(r chi.Router) {
		r.Get("/", toolHandler.GetAllToolManifests)
		r.Post("/call", toolHandler.CallTool)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))

	return r
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

package api

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
	"github.com/nathanaday/iot-data-sandbox/internal/storage"
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

func SetupRouter(store *persistence.Store, fileStore *storage.FileStore) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(corsMiddleware)

	dataSourceHandler := NewDataSourceHandler(store, fileStore)
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

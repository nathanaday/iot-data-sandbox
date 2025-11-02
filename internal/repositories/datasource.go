package repositories

import "github.com/nathanaday/iot-data-sandbox/internal/schemas"

// DataSourceRepository defines the interface for datasource persistence operations
type DataSourceRepository interface {
	// Save inserts or updates a datasource
	Save(ds *schemas.DataSourceSchema) error

	// FindByID retrieves a datasource by ID
	FindByID(id int64) (*schemas.DataSourceSchema, error)

	// FindAll retrieves all datasources
	FindAll() ([]*schemas.DataSourceSchema, error)

	// Delete removes a datasource by ID
	Delete(id int64) error
}

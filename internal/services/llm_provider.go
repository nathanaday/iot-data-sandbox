package services

import (
	"github.com/nathanaday/iot-data-sandbox/internal/models"
	"github.com/nathanaday/iot-data-sandbox/internal/persistence"
)

type LLMProviderService struct {
	store *persistence.Store
}

func NewLLMProviderService(store *persistence.Store) *LLMProviderService {
	return &LLMProviderService{
		store: store,
	}
}

// Create creates a new LLM provider configuration
func (s *LLMProviderService) Create(provider *models.LLMProvider) error {
	return s.store.SaveLLMProvider(provider)
}

// LoadByID retrieves an LLM provider by ID
func (s *LLMProviderService) LoadByID(id int64) (*models.LLMProvider, error) {
	return s.store.LoadLLMProvider(id)
}

// LoadAll retrieves all LLM provider configurations
func (s *LLMProviderService) LoadAll() ([]*models.LLMProvider, error) {
	return s.store.LoadAllLLMProviders()
}

// LoadEnabled retrieves only enabled LLM providers
func (s *LLMProviderService) LoadEnabled() ([]*models.LLMProvider, error) {
	return s.store.LoadEnabledLLMProviders()
}

// LoadByType retrieves all LLM providers of a specific type
func (s *LLMProviderService) LoadByType(providerType models.LLMProviderType) ([]*models.LLMProvider, error) {
	return s.store.LoadLLMProvidersByType(providerType)
}

// Update updates an existing LLM provider configuration
func (s *LLMProviderService) Update(provider *models.LLMProvider) error {
	return s.store.SaveLLMProvider(provider)
}

// Delete removes an LLM provider configuration
func (s *LLMProviderService) Delete(id int64) error {
	return s.store.DeleteLLMProvider(id)
}

// SetEnabled updates the enabled status of an LLM provider
func (s *LLMProviderService) SetEnabled(id int64, enabled bool) error {
	return s.store.SetLLMProviderEnabled(id, enabled)
}

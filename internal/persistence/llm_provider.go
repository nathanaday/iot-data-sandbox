package persistence

import (
	"database/sql"
	"time"

	"github.com/nathanaday/iot-data-sandbox/internal/crypto"
	"github.com/nathanaday/iot-data-sandbox/internal/models"
)

// SaveLLMProvider inserts or updates an LLMProvider
// The API key is automatically encrypted before storage
func (s *Store) SaveLLMProvider(p *models.LLMProvider) error {
	// Encrypt API key before storing
	encryptedKey, err := crypto.EncryptAPIKey(p.APIKey)
	if err != nil {
		return err
	}

	now := time.Now()
	if p.LLMProviderId == 0 {
		p.CreatedAt = now
		p.UpdatedAt = now
		result, err := s.db.Exec(`
            INSERT INTO llm_providers (provider_type, name, api_key_encrypted, base_url, default_model, is_enabled, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			p.ProviderType, p.Name, encryptedKey, p.BaseURL, p.DefaultModel, p.IsEnabled, p.CreatedAt, p.UpdatedAt,
		)
		if err != nil {
			return err
		}
		p.LLMProviderId, _ = result.LastInsertId()
	} else {
		p.UpdatedAt = now
		_, err := s.db.Exec(`
            UPDATE llm_providers
            SET provider_type=?, name=?, api_key_encrypted=?, base_url=?, default_model=?, is_enabled=?, updated_at=?
            WHERE llm_provider_id=?`,
			p.ProviderType, p.Name, encryptedKey, p.BaseURL, p.DefaultModel, p.IsEnabled, p.UpdatedAt, p.LLMProviderId,
		)
		return err
	}
	return nil
}

// LoadLLMProvider retrieves an LLMProvider by ID
// The API key is automatically decrypted when loaded
func (s *Store) LoadLLMProvider(id int64) (*models.LLMProvider, error) {
	p := &models.LLMProvider{}
	var encryptedKey sql.NullString
	var baseURL sql.NullString
	var defaultModel sql.NullString

	err := s.db.QueryRow(`
        SELECT llm_provider_id, provider_type, name, api_key_encrypted, base_url, default_model, is_enabled, created_at, updated_at
        FROM llm_providers WHERE llm_provider_id=?`, id,
	).Scan(&p.LLMProviderId, &p.ProviderType, &p.Name, &encryptedKey, &baseURL, &defaultModel, &p.IsEnabled, &p.CreatedAt, &p.UpdatedAt)

	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if baseURL.Valid {
		p.BaseURL = &baseURL.String
	}
	if defaultModel.Valid {
		p.DefaultModel = &defaultModel.String
	}

	// Decrypt API key
	if encryptedKey.Valid {
		decryptedKey, err := crypto.DecryptAPIKey(encryptedKey.String)
		if err != nil {
			return nil, err
		}
		p.APIKey = decryptedKey
	}

	return p, nil
}

// LoadAllLLMProviders retrieves all LLMProviders ordered by creation date
// API keys are automatically decrypted when loaded
func (s *Store) LoadAllLLMProviders() ([]*models.LLMProvider, error) {
	rows, err := s.db.Query(`
        SELECT llm_provider_id, provider_type, name, api_key_encrypted, base_url, default_model, is_enabled, created_at, updated_at
        FROM llm_providers ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*models.LLMProvider
	for rows.Next() {
		p := &models.LLMProvider{}
		var encryptedKey sql.NullString
		var baseURL sql.NullString
		var defaultModel sql.NullString

		if err := rows.Scan(&p.LLMProviderId, &p.ProviderType, &p.Name, &encryptedKey, &baseURL, &defaultModel, &p.IsEnabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}

		// Handle nullable fields
		if baseURL.Valid {
			p.BaseURL = &baseURL.String
		}
		if defaultModel.Valid {
			p.DefaultModel = &defaultModel.String
		}

		// Decrypt API key
		if encryptedKey.Valid {
			decryptedKey, err := crypto.DecryptAPIKey(encryptedKey.String)
			if err != nil {
				return nil, err
			}
			p.APIKey = decryptedKey
		}

		providers = append(providers, p)
	}
	return providers, rows.Err()
}

// LoadLLMProvidersByType retrieves all LLMProviders of a specific type
func (s *Store) LoadLLMProvidersByType(providerType models.LLMProviderType) ([]*models.LLMProvider, error) {
	rows, err := s.db.Query(`
        SELECT llm_provider_id, provider_type, name, api_key_encrypted, base_url, default_model, is_enabled, created_at, updated_at
        FROM llm_providers WHERE provider_type=? ORDER BY created_at DESC`, providerType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*models.LLMProvider
	for rows.Next() {
		p := &models.LLMProvider{}
		var encryptedKey sql.NullString
		var baseURL sql.NullString
		var defaultModel sql.NullString

		if err := rows.Scan(&p.LLMProviderId, &p.ProviderType, &p.Name, &encryptedKey, &baseURL, &defaultModel, &p.IsEnabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}

		// Handle nullable fields
		if baseURL.Valid {
			p.BaseURL = &baseURL.String
		}
		if defaultModel.Valid {
			p.DefaultModel = &defaultModel.String
		}

		// Decrypt API key
		if encryptedKey.Valid {
			decryptedKey, err := crypto.DecryptAPIKey(encryptedKey.String)
			if err != nil {
				return nil, err
			}
			p.APIKey = decryptedKey
		}

		providers = append(providers, p)
	}
	return providers, rows.Err()
}

// LoadEnabledLLMProviders retrieves all enabled LLMProviders
func (s *Store) LoadEnabledLLMProviders() ([]*models.LLMProvider, error) {
	rows, err := s.db.Query(`
        SELECT llm_provider_id, provider_type, name, api_key_encrypted, base_url, default_model, is_enabled, created_at, updated_at
        FROM llm_providers WHERE is_enabled=1 ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*models.LLMProvider
	for rows.Next() {
		p := &models.LLMProvider{}
		var encryptedKey sql.NullString
		var baseURL sql.NullString
		var defaultModel sql.NullString

		if err := rows.Scan(&p.LLMProviderId, &p.ProviderType, &p.Name, &encryptedKey, &baseURL, &defaultModel, &p.IsEnabled, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}

		// Handle nullable fields
		if baseURL.Valid {
			p.BaseURL = &baseURL.String
		}
		if defaultModel.Valid {
			p.DefaultModel = &defaultModel.String
		}

		// Decrypt API key
		if encryptedKey.Valid {
			decryptedKey, err := crypto.DecryptAPIKey(encryptedKey.String)
			if err != nil {
				return nil, err
			}
			p.APIKey = decryptedKey
		}

		providers = append(providers, p)
	}
	return providers, rows.Err()
}

// DeleteLLMProvider removes an LLMProvider by ID
func (s *Store) DeleteLLMProvider(id int64) error {
	_, err := s.db.Exec("DELETE FROM llm_providers WHERE llm_provider_id=?", id)
	return err
}

// SetLLMProviderEnabled updates the enabled status of an LLMProvider
func (s *Store) SetLLMProviderEnabled(id int64, enabled bool) error {
	_, err := s.db.Exec(`
        UPDATE llm_providers SET is_enabled=?, updated_at=? WHERE llm_provider_id=?`,
		enabled, time.Now(), id,
	)
	return err
}

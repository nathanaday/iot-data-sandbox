package persistence

import (
	"database/sql"

	"github.com/nathanaday/iot-data-sandbox/models"
)

// SaveTool inserts or updates a ToolModel with its auth properties
func (s *Store) SaveTool(tool *models.ToolModel) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if tool.ToolId == 0 {
		// Insert tool
		result, err := tx.Exec(`
            INSERT INTO tools (name, fx_name, timeout_s, is_enabled, when_last_call, 
                             num_calls, max_calls, num_call_reset)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			tool.Name, tool.FxName, tool.TimeoutS, tool.IsEnabled, tool.WhenLastCall,
			tool.NumCalls, tool.MaxCalls, tool.NumCallReset,
		)
		if err != nil {
			return err
		}
		tool.ToolId, _ = result.LastInsertId()
	} else {
		// Update tool
		_, err := tx.Exec(`
            UPDATE tools 
            SET name=?, fx_name=?, timeout_s=?, is_enabled=?, when_last_call=?,
                num_calls=?, max_calls=?, num_call_reset=?
            WHERE tool_id=?`,
			tool.Name, tool.FxName, tool.TimeoutS, tool.IsEnabled, tool.WhenLastCall,
			tool.NumCalls, tool.MaxCalls, tool.NumCallReset, tool.ToolId,
		)
		if err != nil {
			return err
		}
	}

	// Save auth props if they exist
	if tool.AuthProps != nil {
		tool.AuthProps.ToolId = tool.ToolId
		_, err := tx.Exec(`
            INSERT OR REPLACE INTO tool_auth_props 
            (tool_id, hashed_api_key, hashed_username, hashed_password)
            VALUES (?, ?, ?, ?)`,
			tool.AuthProps.ToolId, tool.AuthProps.HashedApiKey,
			tool.AuthProps.HashedUsername, tool.AuthProps.HashedPassword,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// LoadTool retrieves a ToolModel by ID including auth properties
func (s *Store) LoadTool(id int64) (*models.ToolModel, error) {
	tool := &models.ToolModel{}
	err := s.db.QueryRow(`
        SELECT tool_id, name, fx_name, timeout_s, is_enabled, when_last_call,
               num_calls, max_calls, num_call_reset
        FROM tools WHERE tool_id=?`, id,
	).Scan(&tool.ToolId, &tool.Name, &tool.FxName, &tool.TimeoutS, &tool.IsEnabled,
		&tool.WhenLastCall, &tool.NumCalls, &tool.MaxCalls, &tool.NumCallReset)

	if err != nil {
		return nil, err
	}

	// Load auth props if they exist
	authProps := &models.ToolAuthProps{}
	err = s.db.QueryRow(`
        SELECT tool_id, hashed_api_key, hashed_username, hashed_password
        FROM tool_auth_props WHERE tool_id=?`, id,
	).Scan(&authProps.ToolId, &authProps.HashedApiKey,
		&authProps.HashedUsername, &authProps.HashedPassword)

	if err == nil {
		tool.AuthProps = authProps
	} else if err != sql.ErrNoRows {
		return nil, err
	}

	return tool, nil
}

// LoadEnabledTools retrieves all enabled ToolModels
func (s *Store) LoadEnabledTools() ([]*models.ToolModel, error) {
	rows, err := s.db.Query(`
        SELECT tool_id, name, fx_name, timeout_s, is_enabled, when_last_call,
               num_calls, max_calls, num_call_reset
        FROM tools WHERE is_enabled=1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []*models.ToolModel
	for rows.Next() {
		tool := &models.ToolModel{}
		if err := rows.Scan(&tool.ToolId, &tool.Name, &tool.FxName, &tool.TimeoutS,
			&tool.IsEnabled, &tool.WhenLastCall, &tool.NumCalls,
			&tool.MaxCalls, &tool.NumCallReset); err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}
	return tools, rows.Err()
}

// DeleteTool removes a ToolModel by ID (auth props cascade delete automatically)
func (s *Store) DeleteTool(id int64) error {
	// Auth props will be deleted automatically due to CASCADE
	_, err := s.db.Exec("DELETE FROM tools WHERE tool_id=?", id)
	return err
}


package persistence

import "github.com/nathanaday/iot-data-sandbox/internal/schemas"

// SaveProject inserts or updates a Project
func (s *Store) SaveProject(p *schemas.ProjectSchema) error {
	if p.ProjectId == 0 {
		result, err := s.db.Exec(`
            INSERT INTO projects (name, when_created)
            VALUES (?, ?)`,
			p.Name, p.WhenCreated,
		)
		if err != nil {
			return err
		}
		p.ProjectId, _ = result.LastInsertId()
	} else {
		_, err := s.db.Exec(`
            UPDATE projects
            SET name=?, when_created=?
            WHERE project_id=?`,
			p.Name, p.WhenCreated, p.ProjectId,
		)
		return err
	}
	return nil
}

// LoadProject retrieves a Project by ID
func (s *Store) LoadProject(id int64) (*schemas.ProjectSchema, error) {
	p := &schemas.ProjectSchema{}
	err := s.db.QueryRow(`
        SELECT project_id, name, when_created
        FROM projects WHERE project_id=?`, id,
	).Scan(&p.ProjectId, &p.Name, &p.WhenCreated)

	if err != nil {
		return nil, err
	}
	return p, nil
}

// LoadAllProjects retrieves all Projects ordered by creation date
func (s *Store) LoadAllProjects() ([]*schemas.ProjectSchema, error) {
	rows, err := s.db.Query(`
        SELECT project_id, name, when_created
        FROM projects ORDER BY when_created DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*schemas.ProjectSchema
	for rows.Next() {
		p := &schemas.ProjectSchema{}
		if err := rows.Scan(&p.ProjectId, &p.Name, &p.WhenCreated); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// DeleteProject removes a Project by ID
// Note: This will cascade delete all data_layers associated with this project
// and set data_sources.project_id to NULL for associated data sources
func (s *Store) DeleteProject(id int64) error {
	_, err := s.db.Exec("DELETE FROM projects WHERE project_id=?", id)
	return err
}

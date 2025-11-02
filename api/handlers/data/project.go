package data

import "github.com/nathanaday/iot-data-sandbox/internal/models"

// Load from SQLite
func LoadProject(projectId int64) (*models.Project, error) {
	project := &models.Project{}
	project.ProjectId = projectId
	return project, nil
}

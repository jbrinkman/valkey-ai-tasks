package models

import (
	"time"
)

// Project represents a collection of related tasks
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewProject creates a new project with the given name and description
func NewProject(id, name, description string) *Project {
	now := time.Now()
	return &Project{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ToMap converts the project to a map for storage in Valkey
func (p *Project) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":          p.ID,
		"name":        p.Name,
		"description": p.Description,
		"created_at":  p.CreatedAt.Format(time.RFC3339),
		"updated_at":  p.UpdatedAt.Format(time.RFC3339),
	}
}

// FromMap populates a project from a map retrieved from Valkey
func (p *Project) FromMap(data map[string]string) error {
	p.ID = data["id"]
	p.Name = data["name"]
	p.Description = data["description"]

	createdAt, err := time.Parse(time.RFC3339, data["created_at"])
	if err != nil {
		return err
	}
	p.CreatedAt = createdAt

	updatedAt, err := time.Parse(time.RFC3339, data["updated_at"])
	if err != nil {
		return err
	}
	p.UpdatedAt = updatedAt

	return nil
}

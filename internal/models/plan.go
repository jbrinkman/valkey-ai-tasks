package models

import (
	"time"
)

// PlanStatus represents the current status of a plan
type PlanStatus string

const (
	PlanStatusNew       PlanStatus = "new"
	PlanStatusInProgress PlanStatus = "inprogress"
	PlanStatusCompleted  PlanStatus = "completed"
	PlanStatusCancelled  PlanStatus = "cancelled"
)

// Plan represents a collection of related tasks
type Plan struct {
	ID            string     `json:"id"`
	ApplicationID string     `json:"application_id"` // Added field for application association
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Status        PlanStatus `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// NewPlan creates a new plan with the given name and description
func NewPlan(id, applicationID, name, description string) *Plan {
	now := time.Now()
	return &Plan{
		ID:            id,
		ApplicationID: applicationID,
		Name:          name,
		Description:   description,
		Status:        PlanStatusNew,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// ToMap converts the plan to a map for storage in Valkey
func (p *Plan) ToMap() map[string]string {
	return map[string]string{
		"id":             p.ID,
		"application_id": p.ApplicationID,
		"name":           p.Name,
		"description":    p.Description,
		"status":         string(p.Status),
		"created_at":     p.CreatedAt.Format(time.RFC3339),
		"updated_at":     p.UpdatedAt.Format(time.RFC3339),
	}
}

// FromMap populates a plan from a map retrieved from Valkey
func (p *Plan) FromMap(data map[string]string) error {
	p.ID = data["id"]
	p.ApplicationID = data["application_id"]
	p.Name = data["name"]
	p.Description = data["description"]
	
	// Handle status with backward compatibility
	if status, ok := data["status"]; ok {
		p.Status = PlanStatus(status)
	} else {
		// Default to "new" for plans without status
		p.Status = PlanStatusNew
	}

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

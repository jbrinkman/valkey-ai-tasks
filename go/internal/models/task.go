package models

import (
	"fmt"
	"time"
)

// TaskStatus represents the current status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// TaskPriority represents the priority level of a task
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
)

// Task represents an individual task within a project
type Task struct {
	ID          string       `json:"id"`
	ProjectID   string       `json:"project_id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	Order       int          `json:"order"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// NewTask creates a new task with the given details
func NewTask(id, projectID, title, description string, priority TaskPriority) *Task {
	now := time.Now()
	return &Task{
		ID:          id,
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		Status:      TaskStatusPending,
		Priority:    priority,
		Order:       0, // Will be set when added to the project
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ToMap converts the task to a map for storage in Valkey
func (t *Task) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":          t.ID,
		"project_id":  t.ProjectID,
		"title":       t.Title,
		"description": t.Description,
		"status":      string(t.Status),
		"priority":    string(t.Priority),
		"order":       t.Order,
		"created_at":  t.CreatedAt.Format(time.RFC3339),
		"updated_at":  t.UpdatedAt.Format(time.RFC3339),
	}
}

// FromMap populates a task from a map retrieved from Valkey
func (t *Task) FromMap(data map[string]string) error {
	t.ID = data["id"]
	t.ProjectID = data["project_id"]
	t.Title = data["title"]
	t.Description = data["description"]
	t.Status = TaskStatus(data["status"])
	t.Priority = TaskPriority(data["priority"])

	order := 0
	if data["order"] != "" {
		// Convert string to int
		_, err := fmt.Sscanf(data["order"], "%d", &order)
		if err != nil {
			return err
		}
	}
	t.Order = order

	createdAt, err := time.Parse(time.RFC3339, data["created_at"])
	if err != nil {
		return err
	}
	t.CreatedAt = createdAt

	updatedAt, err := time.Parse(time.RFC3339, data["updated_at"])
	if err != nil {
		return err
	}
	t.UpdatedAt = updatedAt

	return nil
}

package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
)

// ProjectRepository handles storage operations for projects
type ProjectRepository struct {
	client *ValkeyClient
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(client *ValkeyClient) *ProjectRepository {
	return &ProjectRepository{
		client: client,
	}
}

// Create adds a new project to the storage
func (r *ProjectRepository) Create(ctx context.Context, name, description string) (*models.Project, error) {
	// Generate a unique ID for the project
	id := uuid.New().String()

	// Create a new project
	project := models.NewProject(id, name, description)

	// Convert project to JSON for storage
	projectJSON, err := json.Marshal(project)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal project: %w", err)
	}

	// Store the project in Valkey
	projectKey := GetProjectKey(id)
	err = r.client.client.HSet(ctx, projectKey, project.ToMap()).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store project: %w", err)
	}

	// Add project ID to the projects list
	err = r.client.client.SAdd(ctx, projectsListKey, id).Err()
	if err != nil {
		// Try to clean up the project if adding to the set fails
		r.client.client.Del(ctx, projectKey)
		return nil, fmt.Errorf("failed to add project to list: %w", err)
	}

	return project, nil
}

// Get retrieves a project by ID
func (r *ProjectRepository) Get(ctx context.Context, id string) (*models.Project, error) {
	// Get the project from Valkey
	projectKey := GetProjectKey(id)
	data, err := r.client.client.HGetAll(ctx, projectKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	// Check if the project exists
	if len(data) == 0 {
		return nil, fmt.Errorf("project not found: %s", id)
	}

	// Convert data to project
	project := &models.Project{}
	err = project.FromMap(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse project data: %w", err)
	}

	return project, nil
}

// Update updates an existing project
func (r *ProjectRepository) Update(ctx context.Context, project *models.Project) error {
	// Check if the project exists
	exists, err := r.client.client.SIsMember(ctx, projectsListKey, project.ID).Result()
	if err != nil {
		return fmt.Errorf("failed to check if project exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("project not found: %s", project.ID)
	}

	// Update the project's updated_at timestamp
	project.UpdatedAt = time.Now()

	// Store the updated project
	projectKey := GetProjectKey(project.ID)
	err = r.client.client.HSet(ctx, projectKey, project.ToMap()).Err()
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

// Delete removes a project and all its tasks
func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	// Check if the project exists
	exists, err := r.client.client.SIsMember(ctx, projectsListKey, id).Result()
	if err != nil {
		return fmt.Errorf("failed to check if project exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("project not found: %s", id)
	}

	// Get the project's tasks
	projectTasksKey := GetProjectTasksKey(id)
	taskIDs, err := r.client.client.ZRange(ctx, projectTasksKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get project tasks: %w", err)
	}

	// Delete all tasks
	for _, taskID := range taskIDs {
		taskKey := GetTaskKey(taskID)
		err = r.client.client.Del(ctx, taskKey).Err()
		if err != nil {
			return fmt.Errorf("failed to delete task %s: %w", taskID, err)
		}
	}

	// Delete the project tasks list
	err = r.client.client.Del(ctx, projectTasksKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete project tasks list: %w", err)
	}

	// Delete the project
	projectKey := GetProjectKey(id)
	err = r.client.client.Del(ctx, projectKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Remove the project from the projects list
	err = r.client.client.SRem(ctx, projectsListKey, id).Err()
	if err != nil {
		return fmt.Errorf("failed to remove project from list: %w", err)
	}

	return nil
}

// List returns all projects
func (r *ProjectRepository) List(ctx context.Context) ([]*models.Project, error) {
	// Get all project IDs
	projectIDs, err := r.client.client.SMembers(ctx, projectsListKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get project list: %w", err)
	}

	projects := make([]*models.Project, 0, len(projectIDs))

	// Get each project
	for _, id := range projectIDs {
		project, err := r.Get(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get project %s: %w", id, err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

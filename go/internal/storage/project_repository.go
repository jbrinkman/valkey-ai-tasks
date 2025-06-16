package storage

import (
	"context"
	"fmt"
	"time"

	uuid "github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/valkey-io/valkey-glide/go/v2/options"
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
func (r *ProjectRepository) Create(ctx context.Context, applicationID, name, description string) (*models.Project, error) {
	// Generate a unique ID for the project
	id := uuid.New().String()

	// Create a new project
	project := models.NewProject(id, applicationID, name, description)

	// Store the project in Valkey
	projectKey := GetProjectKey(id)
	_, err := r.client.client.HSet(ctx, projectKey, project.ToMap())
	if err != nil {
		return nil, fmt.Errorf("failed to store project: %w", err)
	}

	// Add project ID to the projects list
	_, err = r.client.client.SAdd(ctx, projectsListKey, []string{id})
	if err != nil {
		// Try to clean up the project if adding to the set fails
		r.client.client.Del(ctx, []string{projectKey})
		return nil, fmt.Errorf("failed to add project to list: %w", err)
	}

	// Add project ID to the application-specific projects list
	appProjectsKey := fmt.Sprintf("app:%s:projects", applicationID)
	_, err = r.client.client.SAdd(ctx, appProjectsKey, []string{id})
	if err != nil {
		// Log the error but don't fail the operation
		fmt.Printf("Warning: failed to add project to application list: %v\n", err)
	}

	return project, nil
}

// Get retrieves a project by ID
func (r *ProjectRepository) Get(ctx context.Context, id string) (*models.Project, error) {
	// Get the project from Valkey
	projectKey := GetProjectKey(id)
	data, err := r.client.client.HGetAll(ctx, projectKey)
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
	exists, err := r.client.client.SIsMember(ctx, projectsListKey, project.ID)
	if err != nil {
		return fmt.Errorf("failed to get result: %w", err)
	}

	if !exists {
		return fmt.Errorf("project not found: %s", project.ID)
	}

	// Update the project's updated_at timestamp
	project.UpdatedAt = time.Now()

	// Store the updated project
	projectKey := GetProjectKey(project.ID)
	_, err = r.client.client.HSet(ctx, projectKey, project.ToMap())
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

// Delete removes a project and all its tasks
func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	// Check if the project exists
	exists, err := r.client.client.SIsMember(ctx, projectsListKey, id)
	if err != nil {
		return fmt.Errorf("failed to get result: %w", err)
	}

	if !exists {
		return fmt.Errorf("project not found: %s", id)
	}

	// Get the project's tasks
	projectTasksKey := GetProjectTasksKey(id)
	opts := options.NewRangeByIndexQuery(0, -1)
	taskIDs, err := r.client.client.ZRange(ctx, projectTasksKey, opts)
	if err != nil {
		return fmt.Errorf("failed to get result: %w", err)
	}

	// Delete all tasks
	for _, taskID := range taskIDs {
		taskKey := GetTaskKey(taskID)
		_, err := r.client.client.Del(ctx, []string{taskKey})
		if err != nil {
			return fmt.Errorf("failed to delete task %s: %w", taskID, err)
		}
	}

	// Delete the project's tasks list
	_, err = r.client.client.Del(ctx, []string{projectTasksKey})
	if err != nil {
		return fmt.Errorf("failed to delete project tasks list: %w", err)
	}

	// Delete the project
	projectKey := GetProjectKey(id)
	_, err = r.client.client.Del(ctx, []string{projectKey})
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Remove project from the projects list
	_, err = r.client.client.SRem(ctx, projectsListKey, []string{id})
	if err != nil {
		return fmt.Errorf("failed to remove project from list: %w", err)
	}

	// Remove project from the application-specific projects list
	appProjectsKey := fmt.Sprintf("app:%s:projects", id)
	_, err = r.client.client.SRem(ctx, appProjectsKey, []string{id})
	if err != nil {
		// Log the error but don't fail the operation
		fmt.Printf("Warning: failed to remove project from application list: %v\n", err)
	}

	return nil
}

// List returns all projects
func (r *ProjectRepository) List(ctx context.Context) ([]*models.Project, error) {
	// Get all project IDs
	projectIDs, err := r.client.client.SMembers(ctx, projectsListKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get project list: %w", err)
	}

	projects := make([]*models.Project, 0, len(projectIDs))

	// Get each project
	for id, _ := range projectIDs {
		project, err := r.Get(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get project %s: %w", id, err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// ListByApplication retrieves all projects for a specific application
func (r *ProjectRepository) ListByApplication(ctx context.Context, applicationID string) ([]*models.Project, error) {
	// Get the list of project IDs for this application
	appProjectsKey := fmt.Sprintf("app:%s:projects", applicationID)
	projectIDs, err := r.client.client.SMembers(ctx, appProjectsKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get application project IDs: %w", err)
	}

	// If there are no projects, return an empty slice
	if len(projectIDs) == 0 {
		return []*models.Project{}, nil
	}

	projects := make([]*models.Project, 0, len(projectIDs))

	// Retrieve each project
	for _, id := range projectIDs {
		projectKey := GetProjectKey(id)
		projectData, err := r.client.client.HGetAll(ctx, projectKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get project %s: %w", id, err)
		}

		// Skip if the project doesn't exist
		if len(projectData) == 0 {
			continue
		}

		// Create a project from the data
		project := &models.Project{}
		if err := project.FromMap(projectData); err != nil {
			return nil, fmt.Errorf("failed to parse project %s: %w", id, err)
		}

		projects = append(projects, project)
	}

	return projects, nil
}

package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/internal/storage"
)

// MockProjectRepository is an in-memory implementation of ProjectRepositoryInterface for testing
type MockProjectRepository struct {
	mu       sync.RWMutex
	projects map[string]*models.Project
}

// NewMockProjectRepository creates a new mock project repository
func NewMockProjectRepository() *MockProjectRepository {
	return &MockProjectRepository{
		projects: make(map[string]*models.Project),
	}
}

// Create adds a new project to the mock storage
func (r *MockProjectRepository) Create(ctx context.Context, applicationID, name, description string) (*models.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.New().String()
	project := models.NewProject(id, applicationID, name, description)
	r.projects[id] = project
	return project, nil
}

// Get retrieves a project from the mock storage
func (r *MockProjectRepository) Get(ctx context.Context, id string) (*models.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	project, exists := r.projects[id]
	if !exists {
		return nil, fmt.Errorf("project not found: %s", id)
	}
	return project, nil
}

// Update modifies an existing project in the mock storage
func (r *MockProjectRepository) Update(ctx context.Context, project *models.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.projects[project.ID]
	if !exists {
		return fmt.Errorf("project not found: %s", project.ID)
	}

	// Update the project's updated_at timestamp
	project.UpdatedAt = time.Now()

	// Store the updated project
	r.projects[project.ID] = project
	return nil
}

// Delete removes a project from the mock storage
func (r *MockProjectRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.projects[id]; !exists {
		return fmt.Errorf("project not found: %s", id)
	}
	delete(r.projects, id)
	return nil
}

// List returns all projects from the mock storage
func (r *MockProjectRepository) List(ctx context.Context) ([]*models.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	projects := make([]*models.Project, 0, len(r.projects))
	for _, project := range r.projects {
		projects = append(projects, project)
	}
	return projects, nil
}

// ListByApplication returns projects for a specific application from the mock storage
func (r *MockProjectRepository) ListByApplication(ctx context.Context, applicationID string) ([]*models.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	projects := make([]*models.Project, 0)
	for _, project := range r.projects {
		if project.ApplicationID == applicationID {
			projects = append(projects, project)
		}
	}
	return projects, nil
}

// Ensure MockProjectRepository implements the ProjectRepositoryInterface
var _ storage.ProjectRepositoryInterface = (*MockProjectRepository)(nil)

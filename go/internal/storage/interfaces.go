package storage

import (
	"context"

	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
)

// ProjectRepositoryInterface defines the interface for project storage operations
type ProjectRepositoryInterface interface {
	Create(ctx context.Context, applicationID, name, description string) (*models.Project, error)
	Get(ctx context.Context, id string) (*models.Project, error)
	Update(ctx context.Context, project *models.Project) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*models.Project, error)
	ListByApplication(ctx context.Context, applicationID string) ([]*models.Project, error)
}

// TaskRepositoryInterface defines the interface for task storage operations
type TaskRepositoryInterface interface {
	Create(ctx context.Context, projectID, title, description string, priority models.TaskPriority) (*models.Task, error)
	CreateBulk(ctx context.Context, projectID string, tasks []TaskCreateInput) ([]*models.Task, error)
	Get(ctx context.Context, id string) (*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id string) error
	ListByProject(ctx context.Context, projectID string) ([]*models.Task, error)
	ListByStatus(ctx context.Context, status models.TaskStatus) ([]*models.Task, error)
}

// Ensure the concrete types implement the interfaces
var _ ProjectRepositoryInterface = (*ProjectRepository)(nil)
var _ TaskRepositoryInterface = (*TaskRepository)(nil)

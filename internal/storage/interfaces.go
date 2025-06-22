package storage

import (
	"context"

	"github.com/jbrinkman/valkey-ai-tasks/internal/models"
)

// PlanRepositoryInterface defines the interface for plan storage operations
type PlanRepositoryInterface interface {
	Create(ctx context.Context, applicationID, name, description string) (*models.Plan, error)
	Get(ctx context.Context, id string) (*models.Plan, error)
	Update(ctx context.Context, plan *models.Plan) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*models.Plan, error)
	ListByApplication(ctx context.Context, applicationID string) ([]*models.Plan, error)
	ListByStatus(ctx context.Context, status models.PlanStatus) ([]*models.Plan, error)
}

// ProjectRepositoryInterface defines the legacy interface for project storage operations
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
	Create(ctx context.Context, planID, title, description string, priority models.TaskPriority) (*models.Task, error)
	CreateBulk(ctx context.Context, planID string, tasks []TaskCreateInput) ([]*models.Task, error)
	Get(ctx context.Context, id string) (*models.Task, error)
	Update(ctx context.Context, task *models.Task) error
	Delete(ctx context.Context, id string) error
	ListByPlan(ctx context.Context, planID string) ([]*models.Task, error)
	ListByStatus(ctx context.Context, status models.TaskStatus) ([]*models.Task, error)
	ListByPlanAndStatus(ctx context.Context, planID string, status models.TaskStatus) ([]*models.Task, error)
	ReorderTask(ctx context.Context, taskID string, newOrder int) error
	ListOrphanedTasks(ctx context.Context) ([]*models.Task, error)
}

// Ensure the concrete types implement the interfaces
var _ PlanRepositoryInterface = (*PlanRepository)(nil)
var _ TaskRepositoryInterface = (*TaskRepository)(nil)

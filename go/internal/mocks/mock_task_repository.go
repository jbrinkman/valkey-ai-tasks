package mocks

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/storage"
)

// MockTaskRepository is an in-memory implementation of TaskRepositoryInterface for testing
type MockTaskRepository struct {
	mu           sync.RWMutex
	tasks        map[string]*models.Task
	projectTasks map[string][]string // projectID -> []taskID
	projectRepo  *MockProjectRepository
}

// NewMockTaskRepository creates a new mock task repository
func NewMockTaskRepository(projectRepo *MockProjectRepository) *MockTaskRepository {
	return &MockTaskRepository{
		tasks:        make(map[string]*models.Task),
		projectTasks: make(map[string][]string),
		projectRepo:  projectRepo,
	}
}

// Create adds a new task to a project in the mock storage
func (r *MockTaskRepository) Create(ctx context.Context, projectID, title, description string, priority models.TaskPriority) (*models.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if project exists
	if r.projectRepo != nil {
		_, err := r.projectRepo.Get(ctx, projectID)
		if err != nil {
			return nil, err
		}
	}

	id := uuid.New().String()
	task := models.NewTask(id, projectID, title, description, priority)

	// Set the order to be the last in the project
	taskIDs, exists := r.projectTasks[projectID]
	if exists {
		task.Order = len(taskIDs)
	} else {
		task.Order = 0
	}

	r.tasks[id] = task
	r.projectTasks[projectID] = append(r.projectTasks[projectID], id)
	return task, nil
}

// CreateBulk adds multiple tasks to a project in the mock storage
func (r *MockTaskRepository) CreateBulk(ctx context.Context, projectID string, taskInputs []storage.TaskCreateInput) ([]*models.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if project exists
	if r.projectRepo != nil {
		_, err := r.projectRepo.Get(ctx, projectID)
		if err != nil {
			return nil, err
		}
	}

	// Get the current order offset
	orderOffset := 0
	taskIDs, exists := r.projectTasks[projectID]
	if exists {
		orderOffset = len(taskIDs)
	}

	// Create all tasks
	createdTasks := make([]*models.Task, 0, len(taskInputs))
	for i, input := range taskInputs {
		id := uuid.New().String()

		// Apply defaults if needed
		description := input.Description
		if description == "" {
			description = "no description provided"
		}

		status := input.Status
		if status == "" {
			status = models.TaskStatusPending
		}

		priority := input.Priority
		if priority == "" {
			priority = models.TaskPriorityMedium
		}

		task := &models.Task{
			ID:          id,
			ProjectID:   projectID,
			Title:       input.Title,
			Description: description,
			Status:      status,
			Priority:    priority,
			Order:       orderOffset + i,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		r.tasks[id] = task
		r.projectTasks[projectID] = append(r.projectTasks[projectID], id)
		createdTasks = append(createdTasks, task)
	}

	return createdTasks, nil
}

// Get retrieves a task from the mock storage
func (r *MockTaskRepository) Get(ctx context.Context, id string) (*models.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, exists := r.tasks[id]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	return task, nil
}

// Update modifies an existing task in the mock storage
func (r *MockTaskRepository) Update(ctx context.Context, task *models.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.tasks[task.ID]
	if !exists {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	// Update the task's updated_at timestamp
	task.UpdatedAt = time.Now()

	// Store the updated task
	r.tasks[task.ID] = task
	return nil
}

// Delete removes a task from the mock storage
func (r *MockTaskRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, exists := r.tasks[id]
	if !exists {
		return fmt.Errorf("task not found: %s", id)
	}

	// Remove from tasks map
	delete(r.tasks, id)

	// Remove from projectTasks map
	projectID := task.ProjectID
	taskIDs := r.projectTasks[projectID]
	for i, taskID := range taskIDs {
		if taskID == id {
			r.projectTasks[projectID] = append(taskIDs[:i], taskIDs[i+1:]...)
			break
		}
	}

	// Update order for remaining tasks in the project
	for _, taskID := range r.projectTasks[projectID] {
		t := r.tasks[taskID]
		if t.Order > task.Order {
			t.Order--
			r.tasks[taskID] = t
		}
	}

	return nil
}

// ListByProject returns all tasks for a specific project from the mock storage
func (r *MockTaskRepository) ListByProject(ctx context.Context, projectID string) ([]*models.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	taskIDs, exists := r.projectTasks[projectID]
	if !exists {
		return []*models.Task{}, nil
	}

	tasks := make([]*models.Task, 0, len(taskIDs))
	for _, id := range taskIDs {
		if task, ok := r.tasks[id]; ok {
			tasks = append(tasks, task)
		}
	}

	// Sort by order
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Order < tasks[j].Order
	})

	return tasks, nil
}

// ListByStatus returns all tasks with a specific status from the mock storage
func (r *MockTaskRepository) ListByStatus(ctx context.Context, status models.TaskStatus) ([]*models.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]*models.Task, 0)
	for _, task := range r.tasks {
		if task.Status == status {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

// Reorder changes the order of a task within its project
func (r *MockTaskRepository) Reorder(ctx context.Context, id string, newOrder int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, exists := r.tasks[id]
	if !exists {
		return fmt.Errorf("task not found: %s", id)
	}

	projectID := task.ProjectID
	taskIDs, exists := r.projectTasks[projectID]
	if !exists {
		return fmt.Errorf("project tasks not found: %s", projectID)
	}

	if newOrder < 0 || newOrder >= len(taskIDs) {
		return fmt.Errorf("invalid order: %d (must be between 0 and %d)", newOrder, len(taskIDs)-1)
	}

	oldOrder := task.Order
	if oldOrder == newOrder {
		return nil // No change needed
	}

	// Update orders for affected tasks
	for _, taskID := range taskIDs {
		t := r.tasks[taskID]
		if oldOrder < newOrder {
			// Moving down: decrement tasks between old and new
			if t.Order > oldOrder && t.Order <= newOrder {
				t.Order--
				r.tasks[taskID] = t
			}
		} else {
			// Moving up: increment tasks between new and old
			if t.Order >= newOrder && t.Order < oldOrder {
				t.Order++
				r.tasks[taskID] = t
			}
		}
	}

	// Set new order for the task
	task.Order = newOrder
	r.tasks[id] = task

	return nil
}

// Ensure MockTaskRepository implements the TaskRepositoryInterface
var _ storage.TaskRepositoryInterface = (*MockTaskRepository)(nil)

package mocks

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/internal/storage"
)

// MockTaskRepository is an in-memory implementation of TaskRepositoryInterface for testing
type MockTaskRepository struct {
	mu        sync.RWMutex
	tasks     map[string]*models.Task
	planTasks map[string][]string // planID -> []taskID
	planRepo  *MockPlanRepository
}

// NewMockTaskRepository creates a new mock task repository
func NewMockTaskRepository(planRepo *MockPlanRepository) *MockTaskRepository {
	return &MockTaskRepository{
		tasks:     make(map[string]*models.Task),
		planTasks: make(map[string][]string),
		planRepo:  planRepo,
	}
}

// Create adds a new task to a plan in the mock storage
func (r *MockTaskRepository) Create(ctx context.Context, planID string, title, description string, priority models.TaskPriority) (*models.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if the plan exists
	if r.planRepo != nil {
		_, err := r.planRepo.Get(ctx, planID)
		if err != nil {
			return nil, err
		}
	}

	id := uuid.New().String()
	task := models.NewTask(id, planID, title, description, priority)

	// Set the order to be the last
	if _, exists := r.planTasks[planID]; !exists {
		r.planTasks[planID] = []string{}
	}
	r.planTasks[planID] = append(r.planTasks[planID], id)
	r.tasks[id] = task
	return task, nil
}

// CreateBulk adds multiple tasks to a plan in the mock storage
func (r *MockTaskRepository) CreateBulk(ctx context.Context, planID string, taskInputs []storage.TaskCreateInput) ([]*models.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if plan exists
	if r.planRepo != nil {
		_, err := r.planRepo.Get(ctx, planID)
		if err != nil {
			return nil, err
		}
	}

	// Get the current order offset
	orderOffset := 0
	taskIDs, exists := r.planTasks[planID]
	if exists {
		orderOffset = len(taskIDs)
	} else {
		r.planTasks[planID] = []string{}
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
			PlanID:      planID,
			Title:       input.Title,
			Description: description,
			Status:      status,
			Priority:    priority,
			Order:       orderOffset + i,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		r.tasks[id] = task
		r.planTasks[planID] = append(r.planTasks[planID], id)
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

	// Remove from plan's task list
	planID := task.PlanID
	taskIDs, exists := r.planTasks[planID]
	if exists {
		newTaskIDs := []string{}
		for _, taskID := range taskIDs {
			if taskID != id {
				newTaskIDs = append(newTaskIDs, taskID)
			}
		}
		r.planTasks[planID] = newTaskIDs
	}

	// Update order for remaining tasks in the plan
	for _, taskID := range r.planTasks[planID] {
		t := r.tasks[taskID]
		if t.Order > task.Order {
			t.Order--
			r.tasks[taskID] = t
		}
	}

	return nil
}

// ListByPlan returns all tasks for a specific plan from the mock storage
func (r *MockTaskRepository) ListByPlan(ctx context.Context, planID string) ([]*models.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	taskIDs, ok := r.planTasks[planID]
	if !ok {
		return []*models.Task{}, nil
	}

	tasks := make([]*models.Task, 0, len(taskIDs))
	for _, id := range taskIDs {
		if task, ok := r.tasks[id]; ok {
			tasks = append(tasks, task)
		}
	}

	// Sort tasks by order
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

// ReorderTask changes the order of a task within its plan
func (r *MockTaskRepository) ReorderTask(ctx context.Context, id string, newOrder int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, exists := r.tasks[id]
	if !exists {
		return fmt.Errorf("task not found: %s", id)
	}

	planID := task.PlanID
	taskIDs, exists := r.planTasks[planID]
	if !exists {
		return fmt.Errorf("project tasks not found: %s", planID)
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

// ListByPlanAndStatus returns all tasks for a plan with the given status
func (r *MockTaskRepository) ListByPlanAndStatus(ctx context.Context, planID string, status models.TaskStatus) ([]*models.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check if the plan exists
	if r.planRepo != nil {
		_, err := r.planRepo.Get(ctx, planID)
		if err != nil {
			return nil, err
		}
	}

	// Get all tasks for the plan
	taskIDs, exists := r.planTasks[planID]
	if !exists {
		return []*models.Task{}, nil
	}

	// Filter tasks by status
	var filteredTasks []*models.Task
	for _, id := range taskIDs {
		task, exists := r.tasks[id]
		if exists && task.Status == status {
			filteredTasks = append(filteredTasks, task)
		}
	}

	// Sort by order
	sort.Slice(filteredTasks, func(i, j int) bool {
		return filteredTasks[i].Order < filteredTasks[j].Order
	})

	return filteredTasks, nil
}

// ListOrphanedTasks returns all tasks that reference non-existent plans
func (r *MockTaskRepository) ListOrphanedTasks(ctx context.Context) ([]*models.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var orphanedTasks []*models.Task

	// Get all existing plan IDs
	existingPlans := make(map[string]bool)
	if r.planRepo != nil {
		for planID := range r.planRepo.plans {
			existingPlans[planID] = true
		}
	}

	// Find tasks with non-existent plan IDs
	for _, task := range r.tasks {
		if task.PlanID != "" && !existingPlans[task.PlanID] {
			// Create a copy of the task
			taskCopy := *task
			orphanedTasks = append(orphanedTasks, &taskCopy)
		}
	}

	return orphanedTasks, nil
}

// Ensure MockTaskRepository implements the TaskRepositoryInterface
var _ storage.TaskRepositoryInterface = (*MockTaskRepository)(nil)

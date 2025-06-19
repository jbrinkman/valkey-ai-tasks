package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/valkey-io/valkey-glide/go/v2/options"
)

// TaskRepository handles storage operations for tasks
type TaskRepository struct {
	client *ValkeyClient
}

// TaskCreateInput represents the input data for creating a task
type TaskCreateInput struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Status      models.TaskStatus   `json:"status"`
	Priority    models.TaskPriority `json:"priority"`
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(client *ValkeyClient) *TaskRepository {
	return &TaskRepository{
		client: client,
	}
}

// Create adds a new task to a plan
func (r *TaskRepository) Create(ctx context.Context, planID, title, description string, priority models.TaskPriority) (*models.Task, error) {
	// Check if the plan exists
	exists, err := r.client.client.SIsMember(ctx, plansListKey, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get result: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("plan not found: %s", planID)
	}

	// Generate a unique ID for the task
	id := uuid.New().String()

	// Create a new task
	task := models.NewTask(id, planID, title, description, priority)

	// Get the next order value for the task
	planTasksKey := GetPlanTasksKey(planID)
	count, err := r.client.client.ZCard(ctx, planTasksKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get task count: %w", err)
	}

	// Set the order to be the last task in the list
	task.Order = int(count)

	// Store the task in Valkey
	taskKey := GetTaskKey(id)
	_, err = r.client.client.HSet(ctx, taskKey, task.ToMap())
	if err != nil {
		return nil, fmt.Errorf("failed to store task: %w", err)
	}

	// Add task to the plan's tasks list with its order as the score
	_, err = r.client.client.ZAdd(ctx, planTasksKey, map[string]float64{id: float64(task.Order)})
	if err != nil {
		// Try to clean up the task if adding to the set fails
		r.client.client.Del(ctx, []string{taskKey})
		return nil, fmt.Errorf("failed to add task to plan: %w", err)
	}

	return task, nil
}

// Get retrieves a task by ID
func (r *TaskRepository) Get(ctx context.Context, id string) (*models.Task, error) {
	// Get the task from Valkey
	taskKey := GetTaskKey(id)
	data, err := r.client.client.HGetAll(ctx, taskKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Check if the task exists
	if len(data) == 0 {
		return nil, fmt.Errorf("task not found: %s", id)
	}

	// Convert data to task
	task := &models.Task{}
	err = task.FromMap(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse task data: %w", err)
	}

	return task, nil
}

// Update updates an existing task
func (r *TaskRepository) Update(ctx context.Context, task *models.Task) error {
	// Check if the task exists
	taskKey := GetTaskKey(task.ID)
	exists, err := r.client.client.Exists(ctx, []string{taskKey})
	if err != nil {
		return fmt.Errorf("failed to check if task exists: %w", err)
	}

	if exists == 0 {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	// Get the current task to check if the plan ID has changed
	currentTask, err := r.Get(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("failed to get current task: %w", err)
	}

	// Update the task's updated_at timestamp
	task.UpdatedAt = time.Now()

	// Store the updated task
	_, err = r.client.client.HSet(ctx, taskKey, task.ToMap())
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// If the plan ID has changed, move the task to the new plan
	if currentTask.PlanID != task.PlanID {
		// Remove from the old plan's tasks list
		oldPlanTasksKey := GetPlanTasksKey(currentTask.PlanID)
		_, err = r.client.client.ZRem(ctx, oldPlanTasksKey, []string{task.ID})
		if err != nil {
			return fmt.Errorf("failed to remove task from old plan: %w", err)
		}

		// Add to the new plan's tasks list
		newPlanTasksKey := GetPlanTasksKey(task.PlanID)
		_, err = r.client.client.ZAdd(ctx, newPlanTasksKey, map[string]float64{task.ID: float64(task.Order)})
		if err != nil {
			return fmt.Errorf("failed to add task to new plan: %w", err)
		}
	}

	return nil
}

// Delete removes a task
func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	// Get the task to find its plan ID
	task, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	// Remove the task from the plan's tasks list
	planTasksKey := GetPlanTasksKey(task.PlanID)
	_, err = r.client.client.ZRem(ctx, planTasksKey, []string{id})
	if err != nil {
		return fmt.Errorf("failed to remove task from plan list: %w", err)
	}

	// Delete the task
	taskKey := GetTaskKey(id)
	_, err = r.client.client.Del(ctx, []string{taskKey})
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Reorder the remaining tasks in the plan
	return r.reorderPlanTasks(ctx, task.PlanID)
}

// ListByPlan returns all tasks for a plan, ordered by their sequence
func (r *TaskRepository) ListByPlan(ctx context.Context, planID string) ([]*models.Task, error) {
	// Check if the plan exists
	exists, err := r.client.client.SIsMember(ctx, plansListKey, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if plan exists: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("plan not found: %s", planID)
	}

	// Get all task IDs for this plan
	planTasksKey := GetPlanTasksKey(planID)
	opts := options.NewRangeByIndexQuery(0, -1)
	taskIDs, err := r.client.client.ZRange(ctx, planTasksKey, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan tasks: %w", err)
	}

	tasks := make([]*models.Task, 0, len(taskIDs))

	// Get each task
	for _, id := range taskIDs {
		task, err := r.Get(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get task %s: %w", id, err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// ListByStatus returns all tasks with the given status
func (r *TaskRepository) ListByStatus(ctx context.Context, status models.TaskStatus) ([]*models.Task, error) {
	// Get all plan IDs
	planIDs, err := r.client.client.SMembers(ctx, plansListKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan list: %w", err)
	}

	var allTasks []*models.Task

	// For each plan, get its tasks and filter by status
	for planID := range planIDs {
		tasks, err := r.ListByPlan(ctx, planID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tasks for plan %s: %w", planID, err)
		}

		// Filter tasks by status
		for _, task := range tasks {
			if task.Status == status {
				allTasks = append(allTasks, task)
			}
		}
	}

	return allTasks, nil
}

// ReorderTask changes the order of a task within its plan
func (r *TaskRepository) ReorderTask(ctx context.Context, taskID string, newOrder int) error {
	// Get the task
	task, err := r.Get(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get all tasks for this plan to reorder them
	tasks, err := r.ListByPlan(ctx, task.PlanID)
	if err != nil {
		return fmt.Errorf("failed to list plan tasks: %w", err)
	}

	// Validate the new order
	if newOrder < 0 || newOrder >= len(tasks) {
		return fmt.Errorf("invalid order: %d (must be between 0 and %d)", newOrder, len(tasks)-1)
	}

	// If the order hasn't changed, do nothing
	if task.Order == newOrder {
		return nil
	}

	// Update the task's order
	task.Order = newOrder
	task.UpdatedAt = time.Now()

	// Store the updated task
	taskKey := GetTaskKey(task.ID)
	_, err = r.client.client.HSet(ctx, taskKey, task.ToMap())
	if err != nil {
		return fmt.Errorf("failed to update task order: %w", err)
	}

	// Reorder all tasks in the plan to ensure they are sequential
	return r.reorderPlanTasks(ctx, task.PlanID)
}

// CreateBulk adds multiple tasks to a plan in a single operation
func (r *TaskRepository) CreateBulk(ctx context.Context, planID string, taskInputs []TaskCreateInput) ([]*models.Task, error) {
	// Check if the plan exists
	exists, err := r.client.client.SIsMember(ctx, plansListKey, planID)
	if err != nil {
		return nil, fmt.Errorf("failed to get result: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("plan not found: %s", planID)
	}

	// Get the next order value for the first task
	planTasksKey := GetPlanTasksKey(planID)
	count, err := r.client.client.ZCard(ctx, planTasksKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get task count: %w", err)
	}

	// Create all tasks
	createdTasks := make([]*models.Task, 0, len(taskInputs))
	for i, input := range taskInputs {
		// Generate a unique ID for the task
		id := uuid.New().String()

		// Set default values if not provided
		priority := input.Priority
		if priority == "" {
			priority = models.TaskPriorityMedium
		}

		status := input.Status
		if status == "" {
			status = models.TaskStatusPending
		}

		description := input.Description
		if description == "" {
			description = "no description provided"
		}

		// Create a new task
		task := models.NewTask(id, planID, input.Title, description, priority)
		task.Status = status
		task.Order = int(count) + i

		// Store the task in Valkey
		taskKey := GetTaskKey(id)
		_, err = r.client.client.HSet(ctx, taskKey, task.ToMap())
		if err != nil {
			// Try to clean up already created tasks
			for _, createdTask := range createdTasks {
				r.client.client.Del(ctx, []string{GetTaskKey(createdTask.ID)})
				r.client.client.ZRem(ctx, planTasksKey, []string{createdTask.ID})
			}
			return nil, fmt.Errorf("failed to store task: %w", err)
		}

		// Add task to the plan's tasks list with its order as the score
		_, err = r.client.client.ZAdd(ctx, planTasksKey, map[string]float64{id: float64(task.Order)})
		if err != nil {
			// Try to clean up the task if adding to the sorted set fails
			r.client.client.Del(ctx, []string{taskKey})
			// Also clean up already created tasks
			for _, createdTask := range createdTasks {
				r.client.client.Del(ctx, []string{GetTaskKey(createdTask.ID)})
				r.client.client.ZRem(ctx, planTasksKey, []string{createdTask.ID})
			}
			return nil, fmt.Errorf("failed to add task to plan: %w", err)
		}

		createdTasks = append(createdTasks, task)
	}

	return createdTasks, nil
}

// reorderPlanTasks updates the order of all tasks in a plan to ensure they are sequential
func (r *TaskRepository) reorderPlanTasks(ctx context.Context, planID string) error {
	// Get all task IDs for the plan, ordered by their score (order)
	planTasksKey := GetPlanTasksKey(planID)
	opts := options.NewRangeByIndexQuery(0, -1)
	taskIDs, err := r.client.client.ZRange(ctx, planTasksKey, opts)
	if err != nil {
		return fmt.Errorf("failed to get plan tasks: %w", err)
	}

	// Update each task's order to match its position in the slice
	for i, taskID := range taskIDs {
		// Get the task
		task, err := r.Get(ctx, taskID)
		if err != nil {
			return fmt.Errorf("failed to get task %s: %w", taskID, err)
		}

		if task.Order != i {
			task.Order = i
			taskKey := GetTaskKey(task.ID)
			_, err := r.client.client.HSet(ctx, taskKey, task.ToMap())
			if err != nil {
				return fmt.Errorf("failed to update task order: %w", err)
			}

			// Update the task's score in the sorted set
			_, err = r.client.client.ZAdd(ctx, planTasksKey, map[string]float64{task.ID: float64(i)})
			if err != nil {
				return fmt.Errorf("failed to update task order in hash: %w", err)
			}
		}
	}

	return nil
}

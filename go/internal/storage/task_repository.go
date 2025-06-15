package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
)

// TaskRepository handles storage operations for tasks
type TaskRepository struct {
	client *ValkeyClient
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(client *ValkeyClient) *TaskRepository {
	return &TaskRepository{
		client: client,
	}
}

// Create adds a new task to a project
func (r *TaskRepository) Create(ctx context.Context, projectID, title, description string, priority models.TaskPriority) (*models.Task, error) {
	// Check if the project exists
	exists, err := r.client.client.SIsMember(ctx, projectsListKey, projectID).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to check if project exists: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("project not found: %s", projectID)
	}

	// Generate a unique ID for the task
	id := uuid.New().String()

	// Create a new task
	task := models.NewTask(id, projectID, title, description, priority)

	// Get the next order value for the task
	projectTasksKey := GetProjectTasksKey(projectID)
	count, err := r.client.client.ZCard(ctx, projectTasksKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get task count: %w", err)
	}

	// Set the order to be the last task in the list
	task.Order = int(count)

	// Store the task in Valkey
	taskKey := GetTaskKey(id)
	err = r.client.client.HSet(ctx, taskKey, task.ToMap()).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store task: %w", err)
	}

	// Add task to the project's tasks list with its order as the score
	err = r.client.client.ZAdd(ctx, projectTasksKey, &valkey.Z{
		Score:  float64(task.Order),
		Member: id,
	}).Err()
	if err != nil {
		// Try to clean up the task if adding to the sorted set fails
		r.client.client.Del(ctx, taskKey)
		return nil, fmt.Errorf("failed to add task to project: %w", err)
	}

	return task, nil
}

// Get retrieves a task by ID
func (r *TaskRepository) Get(ctx context.Context, id string) (*models.Task, error) {
	// Get the task from Valkey
	taskKey := GetTaskKey(id)
	data, err := r.client.client.HGetAll(ctx, taskKey).Result()
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
	exists, err := r.client.client.Exists(ctx, taskKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check if task exists: %w", err)
	}

	if exists == 0 {
		return fmt.Errorf("task not found: %s", task.ID)
	}

	// Get the current task to check if the project ID has changed
	currentTask, err := r.Get(ctx, task.ID)
	if err != nil {
		return fmt.Errorf("failed to get current task: %w", err)
	}

	// Update the task's updated_at timestamp
	task.UpdatedAt = time.Now()

	// Store the updated task
	err = r.client.client.HSet(ctx, taskKey, task.ToMap()).Err()
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	// If the project ID has changed, move the task to the new project
	if currentTask.ProjectID != task.ProjectID {
		// Remove from the old project's tasks list
		oldProjectTasksKey := GetProjectTasksKey(currentTask.ProjectID)
		err = r.client.client.ZRem(ctx, oldProjectTasksKey, task.ID).Err()
		if err != nil {
			return fmt.Errorf("failed to remove task from old project: %w", err)
		}

		// Add to the new project's tasks list
		newProjectTasksKey := GetProjectTasksKey(task.ProjectID)
		count, err := r.client.client.ZCard(ctx, newProjectTasksKey).Result()
		if err != nil {
			return fmt.Errorf("failed to get task count for new project: %w", err)
		}

		// Set the order to be the last task in the new project
		task.Order = int(count)

		// Update the task with the new order
		err = r.client.client.HSet(ctx, taskKey, "order", task.Order).Err()
		if err != nil {
			return fmt.Errorf("failed to update task order: %w", err)
		}

		// Add to the new project's tasks list
		err = r.client.client.ZAdd(ctx, newProjectTasksKey, &valkey.Z{
			Score:  float64(task.Order),
			Member: task.ID,
		}).Err()
		if err != nil {
			return fmt.Errorf("failed to add task to new project: %w", err)
		}
	}

	return nil
}

// Delete removes a task
func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	// Get the task to find its project
	task, err := r.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Remove the task from the project's tasks list
	projectTasksKey := GetProjectTasksKey(task.ProjectID)
	err = r.client.client.ZRem(ctx, projectTasksKey, id).Err()
	if err != nil {
		return fmt.Errorf("failed to remove task from project: %w", err)
	}

	// Delete the task
	taskKey := GetTaskKey(id)
	err = r.client.client.Del(ctx, taskKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Reorder the remaining tasks in the project
	err = r.reorderProjectTasks(ctx, task.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to reorder project tasks: %w", err)
	}

	return nil
}

// ListByProject returns all tasks for a project, ordered by their sequence
func (r *TaskRepository) ListByProject(ctx context.Context, projectID string) ([]*models.Task, error) {
	// Check if the project exists
	exists, err := r.client.client.SIsMember(ctx, projectsListKey, projectID).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to check if project exists: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("project not found: %s", projectID)
	}

	// Get all task IDs for the project, ordered by their score (order)
	projectTasksKey := GetProjectTasksKey(projectID)
	taskIDs, err := r.client.client.ZRange(ctx, projectTasksKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get project tasks: %w", err)
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
	// Get all project IDs
	projectIDs, err := r.client.client.SMembers(ctx, projectsListKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get project list: %w", err)
	}

	var allTasks []*models.Task

	// For each project, get its tasks and filter by status
	for _, projectID := range projectIDs {
		tasks, err := r.ListByProject(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tasks for project %s: %w", projectID, err)
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

// ReorderTask changes the order of a task within its project
func (r *TaskRepository) ReorderTask(ctx context.Context, taskID string, newOrder int) error {
	// Get the task
	task, err := r.Get(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Get all tasks for the project
	projectTasks, err := r.ListByProject(ctx, task.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get project tasks: %w", err)
	}

	// Validate the new order
	if newOrder < 0 || newOrder >= len(projectTasks) {
		return fmt.Errorf("invalid order: %d (must be between 0 and %d)", newOrder, len(projectTasks)-1)
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
	err = r.client.client.HSet(ctx, taskKey, "order", task.Order, "updated_at", task.UpdatedAt.Format(time.RFC3339)).Err()
	if err != nil {
		return fmt.Errorf("failed to update task order: %w", err)
	}

	// Update the task's score in the sorted set
	projectTasksKey := GetProjectTasksKey(task.ProjectID)
	err = r.client.client.ZAdd(ctx, projectTasksKey, &valkey.Z{
		Score:  float64(newOrder),
		Member: task.ID,
	}).Err()
	if err != nil {
		return fmt.Errorf("failed to update task order in project: %w", err)
	}

	// Reorder other tasks in the project
	return r.reorderProjectTasks(ctx, task.ProjectID)
}

// reorderProjectTasks ensures that all tasks in a project have consecutive order values
func (r *TaskRepository) reorderProjectTasks(ctx context.Context, projectID string) error {
	// Get all task IDs for the project, ordered by their score (order)
	projectTasksKey := GetProjectTasksKey(projectID)
	taskIDs, err := r.client.client.ZRange(ctx, projectTasksKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get project tasks: %w", err)
	}

	// Update each task's order to match its position in the list
	for i, id := range taskIDs {
		// Get the current score for the task
		score, err := r.client.client.ZScore(ctx, projectTasksKey, id).Result()
		if err != nil {
			return fmt.Errorf("failed to get task score: %w", err)
		}

		// If the score doesn't match the position, update it
		if int(score) != i {
			// Update the task's order in the sorted set
			err = r.client.client.ZAdd(ctx, projectTasksKey, &valkey.Z{
				Score:  float64(i),
				Member: id,
			}).Err()
			if err != nil {
				return fmt.Errorf("failed to update task order: %w", err)
			}

			// Update the task's order field
			taskKey := GetTaskKey(id)
			err = r.client.client.HSet(ctx, taskKey, "order", i).Err()
			if err != nil {
				return fmt.Errorf("failed to update task order field: %w", err)
			}
		}
	}

	return nil
}

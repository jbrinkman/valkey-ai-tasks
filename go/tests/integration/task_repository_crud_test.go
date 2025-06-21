package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/storage"
	"github.com/jbrinkman/valkey-ai-tasks/go/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskRepositoryCRUD tests the basic CRUD operations for the TaskRepository
func TestTaskRepositoryCRUD(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set up test context
	ctx, _, cleanup := utils.SetupValkeyTest(t)
	defer cleanup()

	// Create Valkey client
	valkeyClient, err := storage.NewValkeyClient("localhost", 6379, "", "")
	require.NoError(t, err, "Failed to create Valkey client")
	defer valkeyClient.Close()

	// Create repositories
	planRepo := storage.NewPlanRepository(valkeyClient)
	taskRepo := storage.NewTaskRepository(valkeyClient)

	// Create a test plan
	appID := "test-app-" + uuid.New().String()
	plan, err := planRepo.Create(ctx, appID, "Test Plan", "Test plan description")
	require.NoError(t, err, "Failed to create test plan")

	// Test: Create a task
	t.Run("CreateTask", func(t *testing.T) {
		task, err := taskRepo.Create(ctx, plan.ID, "Test Task", "Test task description", models.TaskPriorityHigh)
		assert.NoError(t, err, "Failed to create task")
		assert.NotEmpty(t, task.ID, "Task ID should not be empty")
		assert.Equal(t, "Test Task", task.Title, "Task title should match")
		assert.Equal(t, "Test task description", task.Description, "Task description should match")
		assert.Equal(t, models.TaskPriorityHigh, task.Priority, "Task priority should match")
		assert.Equal(t, models.TaskStatusPending, task.Status, "Task should have default pending status")
		assert.Equal(t, 0, task.Order, "First task should have order 0")
		assert.Equal(t, plan.ID, task.PlanID, "Task should be associated with the correct plan")
	})

	// Create another task for subsequent tests
	task, err := taskRepo.Create(ctx, plan.ID, "Test Task", "Test task description", models.TaskPriorityHigh)
	require.NoError(t, err, "Failed to create task for subsequent tests")

	// Test: Get a task
	t.Run("GetTask", func(t *testing.T) {
		retrievedTask, err := taskRepo.Get(ctx, task.ID)
		assert.NoError(t, err, "Failed to get task")
		assert.Equal(t, task.ID, retrievedTask.ID, "Task ID should match")
		assert.Equal(t, task.Title, retrievedTask.Title, "Task title should match")
		assert.Equal(t, task.Description, retrievedTask.Description, "Task description should match")
		assert.Equal(t, task.Priority, retrievedTask.Priority, "Task priority should match")
		assert.Equal(t, task.Status, retrievedTask.Status, "Task status should match")
		assert.Equal(t, task.Order, retrievedTask.Order, "Task order should match")
		assert.Equal(t, task.PlanID, retrievedTask.PlanID, "Task plan ID should match")
	})

	// Test: Get a non-existent task
	t.Run("GetNonExistentTask", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		_, err := taskRepo.Get(ctx, nonExistentID)
		assert.Error(t, err, "Getting non-existent task should return error")
		assert.Contains(t, err.Error(), "task not found", "Error should indicate task not found")
	})

	// Test: Update a task
	t.Run("UpdateTask", func(t *testing.T) {
		// Retrieve the task first
		taskToUpdate, err := taskRepo.Get(ctx, task.ID)
		assert.NoError(t, err, "Failed to get task for update")

		// Update task properties
		taskToUpdate.Title = "Updated Task Title"
		taskToUpdate.Description = "Updated task description"
		taskToUpdate.Priority = models.TaskPriorityLow
		taskToUpdate.Status = models.TaskStatusInProgress

		// Perform the update
		err = taskRepo.Update(ctx, taskToUpdate)
		assert.NoError(t, err, "Failed to update task")

		// Retrieve the task again to verify updates
		updatedTask, err := taskRepo.Get(ctx, task.ID)
		assert.NoError(t, err, "Failed to get updated task")
		assert.Equal(t, "Updated Task Title", updatedTask.Title, "Task title should be updated")
		assert.Equal(t, "Updated task description", updatedTask.Description, "Task description should be updated")
		assert.Equal(t, models.TaskPriorityLow, updatedTask.Priority, "Task priority should be updated")
		assert.Equal(t, models.TaskStatusInProgress, updatedTask.Status, "Task status should be updated")
	})

	// Test: Update a non-existent task
	t.Run("UpdateNonExistentTask", func(t *testing.T) {
		nonExistentTask := &models.Task{
			ID:          uuid.New().String(),
			Title:       "Non-existent Task",
			Description: "This task doesn't exist",
			PlanID:      plan.ID,
		}
		err := taskRepo.Update(ctx, nonExistentTask)
		assert.Error(t, err, "Updating non-existent task should return error")
		assert.Contains(t, err.Error(), "task not found", "Error should indicate task not found")
	})

	// Test: List tasks by plan
	t.Run("ListTasksByPlan", func(t *testing.T) {
		// Create additional tasks for the plan
		task2, err := taskRepo.Create(ctx, plan.ID, "Second Task", "Second task description", models.TaskPriorityMedium)
		assert.NoError(t, err, "Failed to create second task")

		task3, err := taskRepo.Create(ctx, plan.ID, "Third Task", "Third task description", models.TaskPriorityLow)
		assert.NoError(t, err, "Failed to create third task")

		// List tasks for the plan
		tasks, err := taskRepo.ListByPlan(ctx, plan.ID)
		assert.NoError(t, err, "Failed to list tasks by plan")
		assert.Equal(t, 3, len(tasks), "Should have 3 tasks in the plan")

		// Verify tasks are ordered correctly
		assert.Equal(t, 0, tasks[0].Order, "First task should have order 0")
		assert.Equal(t, 1, tasks[1].Order, "Second task should have order 1")
		assert.Equal(t, 2, tasks[2].Order, "Third task should have order 2")

		// Clean up the additional tasks
		err = taskRepo.Delete(ctx, task2.ID)
		assert.NoError(t, err, "Failed to delete second task")

		err = taskRepo.Delete(ctx, task3.ID)
		assert.NoError(t, err, "Failed to delete third task")
	})

	// Test: List tasks by status
	t.Run("ListTasksByStatus", func(t *testing.T) {
		// Create tasks with different statuses
		taskPending, err := taskRepo.Create(ctx, plan.ID, "Pending Task", "A pending task", models.TaskPriorityMedium)
		assert.NoError(t, err, "Failed to create pending task")

		taskInProgress, err := taskRepo.Create(ctx, plan.ID, "In Progress Task", "An in-progress task", models.TaskPriorityMedium)
		assert.NoError(t, err, "Failed to create in-progress task")

		// Update the second task to in-progress status
		taskInProgress.Status = models.TaskStatusInProgress
		err = taskRepo.Update(ctx, taskInProgress)
		assert.NoError(t, err, "Failed to update task status")

		// List pending tasks
		pendingTasks, err := taskRepo.ListByStatus(ctx, models.TaskStatusPending)
		assert.NoError(t, err, "Failed to list pending tasks")

		// Find our specific pending task in the results
		foundPending := false
		for _, t := range pendingTasks {
			if t.ID == taskPending.ID {
				foundPending = true
				break
			}
		}
		assert.True(t, foundPending, "Should find the pending task in pending tasks list")

		// List in-progress tasks
		inProgressTasks, err := taskRepo.ListByStatus(ctx, models.TaskStatusInProgress)
		assert.NoError(t, err, "Failed to list in-progress tasks")

		// Find our specific in-progress task in the results
		foundInProgress := false
		for _, t := range inProgressTasks {
			if t.ID == taskInProgress.ID {
				foundInProgress = true
				break
			}
		}
		assert.True(t, foundInProgress, "Should find the in-progress task in in-progress tasks list")

		// Clean up
		err = taskRepo.Delete(ctx, taskPending.ID)
		assert.NoError(t, err, "Failed to delete pending task")

		err = taskRepo.Delete(ctx, taskInProgress.ID)
		assert.NoError(t, err, "Failed to delete in-progress task")
	})

	// Test: Reorder tasks
	t.Run("ReorderTask", func(t *testing.T) {
		// Create three tasks for reordering
		task1, err := taskRepo.Create(ctx, plan.ID, "Task 1", "First task", models.TaskPriorityMedium)
		assert.NoError(t, err, "Failed to create first task")

		task2, err := taskRepo.Create(ctx, plan.ID, "Task 2", "Second task", models.TaskPriorityMedium)
		assert.NoError(t, err, "Failed to create second task")

		task3, err := taskRepo.Create(ctx, plan.ID, "Task 3", "Third task", models.TaskPriorityMedium)
		assert.NoError(t, err, "Failed to create third task")

		// Verify initial order
		tasks, err := taskRepo.ListByPlan(ctx, plan.ID)
		assert.NoError(t, err, "Failed to list tasks")
		assert.Equal(t, 3, len(tasks), "Should have 3 tasks")

		// Initial order should be: task1 (0), task2 (1), task3 (2)

		// Move task1 to position 2 (last)
		err = taskRepo.ReorderTask(ctx, task1.ID, 2)
		assert.NoError(t, err, "Failed to reorder task")

		// Check the new order
		tasksAfterReorder, err := taskRepo.ListByPlan(ctx, plan.ID)
		assert.NoError(t, err, "Failed to list tasks after reorder")

		// New order should be: task2 (0), task3 (1), task1 (2)
		assert.Equal(t, task2.ID, tasksAfterReorder[0].ID, "First task should now be task2")
		assert.Equal(t, task3.ID, tasksAfterReorder[1].ID, "Second task should now be task3")
		assert.Equal(t, task1.ID, tasksAfterReorder[2].ID, "Third task should now be task1")

		// Verify the order values are updated
		assert.Equal(t, 0, tasksAfterReorder[0].Order, "First task should have order 0")
		assert.Equal(t, 1, tasksAfterReorder[1].Order, "Second task should have order 1")
		assert.Equal(t, 2, tasksAfterReorder[2].Order, "Third task should have order 2")

		// Clean up
		err = taskRepo.Delete(ctx, task1.ID)
		assert.NoError(t, err, "Failed to delete first task")

		err = taskRepo.Delete(ctx, task2.ID)
		assert.NoError(t, err, "Failed to delete second task")

		err = taskRepo.Delete(ctx, task3.ID)
		assert.NoError(t, err, "Failed to delete third task")
	})

	// Test: Delete a task
	t.Run("DeleteTask", func(t *testing.T) {
		// Delete the task
		err := taskRepo.Delete(ctx, task.ID)
		assert.NoError(t, err, "Failed to delete task")

		// Try to get the deleted task
		_, err = taskRepo.Get(ctx, task.ID)
		assert.Error(t, err, "Getting deleted task should return error")
		assert.Contains(t, err.Error(), "task not found", "Error should indicate task not found")
	})

	// Test: Delete a non-existent task
	t.Run("DeleteNonExistentTask", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		err := taskRepo.Delete(ctx, nonExistentID)
		assert.Error(t, err, "Deleting non-existent task should return error")
		assert.Contains(t, err.Error(), "task not found", "Error should indicate task not found")
	})
}

// TestTaskRepositoryEdgeCases tests edge cases for the TaskRepository
func TestTaskRepositoryEdgeCases(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set up test context
	ctx, _, cleanup := utils.SetupValkeyTest(t)
	defer cleanup()

	// Create Valkey client
	valkeyClient, err := storage.NewValkeyClient("localhost", 6379, "", "")
	require.NoError(t, err, "Failed to create Valkey client")
	defer valkeyClient.Close()

	// Create repositories
	planRepo := storage.NewPlanRepository(valkeyClient)
	taskRepo := storage.NewTaskRepository(valkeyClient)

	// Create a test plan
	appID := "test-app-" + uuid.New().String()
	plan, err := planRepo.Create(ctx, appID, "Test Plan", "Test plan description")
	require.NoError(t, err, "Failed to create test plan")

	// Test: Create task with empty title
	t.Run("CreateTaskWithEmptyTitle", func(t *testing.T) {
		task, err := taskRepo.Create(ctx, plan.ID, "", "Task with empty title", models.TaskPriorityMedium)
		assert.NoError(t, err, "Should be able to create task with empty title")
		assert.Empty(t, task.Title, "Task title should be empty")

		// Clean up
		err = taskRepo.Delete(ctx, task.ID)
		assert.NoError(t, err, "Failed to delete task")
	})

	// Test: Create task with empty description
	t.Run("CreateTaskWithEmptyDescription", func(t *testing.T) {
		task, err := taskRepo.Create(ctx, plan.ID, "Task Title", "", models.TaskPriorityMedium)
		assert.NoError(t, err, "Should be able to create task with empty description")
		assert.Empty(t, task.Description, "Task description should be empty")

		// Clean up
		err = taskRepo.Delete(ctx, task.ID)
		assert.NoError(t, err, "Failed to delete task")
	})

	// Test: Create task with non-existent plan
	t.Run("CreateTaskWithNonExistentPlan", func(t *testing.T) {
		nonExistentPlanID := uuid.New().String()
		_, err := taskRepo.Create(ctx, nonExistentPlanID, "Task Title", "Description", models.TaskPriorityMedium)
		assert.Error(t, err, "Creating task with non-existent plan should fail")
		assert.Contains(t, err.Error(), "plan not found", "Error should indicate plan not found")
	})

	// Test: List tasks for non-existent plan
	t.Run("ListTasksForNonExistentPlan", func(t *testing.T) {
		nonExistentPlanID := uuid.New().String()
		_, err := taskRepo.ListByPlan(ctx, nonExistentPlanID)
		assert.Error(t, err, "Listing tasks for non-existent plan should fail")
		assert.Contains(t, err.Error(), "plan not found", "Error should indicate plan not found")
	})

	// Test: Update task with invalid order
	t.Run("ReorderTaskWithInvalidOrder", func(t *testing.T) {
		// Create a task
		task, err := taskRepo.Create(ctx, plan.ID, "Task for reordering", "Description", models.TaskPriorityMedium)
		assert.NoError(t, err, "Failed to create task")

		// Try to reorder with invalid negative order
		err = taskRepo.ReorderTask(ctx, task.ID, -1)
		assert.Error(t, err, "Reordering task with negative order should fail")
		assert.Contains(t, err.Error(), "invalid order", "Error should indicate invalid order")

		// Try to reorder with too large order
		err = taskRepo.ReorderTask(ctx, task.ID, 100)
		assert.Error(t, err, "Reordering task with too large order should fail")
		assert.Contains(t, err.Error(), "invalid order", "Error should indicate invalid order")

		// Clean up
		err = taskRepo.Delete(ctx, task.ID)
		assert.NoError(t, err, "Failed to delete task")
	})

	// Test: Move task between plans
	t.Run("MoveTaskBetweenPlans", func(t *testing.T) {
		// Create a second plan
		secondPlan, err := planRepo.Create(ctx, appID, "Second Plan", "Another plan")
		assert.NoError(t, err, "Failed to create second plan")

		// Create a task in the first plan
		task, err := taskRepo.Create(ctx, plan.ID, "Task to move", "This task will be moved", models.TaskPriorityMedium)
		assert.NoError(t, err, "Failed to create task")

		// Verify task is in first plan
		tasksInFirstPlan, err := taskRepo.ListByPlan(ctx, plan.ID)
		assert.NoError(t, err, "Failed to list tasks in first plan")
		assert.Equal(t, 1, len(tasksInFirstPlan), "First plan should have 1 task")

		// Move task to second plan
		taskToMove, err := taskRepo.Get(ctx, task.ID)
		assert.NoError(t, err, "Failed to get task")

		taskToMove.PlanID = secondPlan.ID
		err = taskRepo.Update(ctx, taskToMove)
		assert.NoError(t, err, "Failed to update task's plan")

		// Verify task is now in second plan
		tasksInFirstPlanAfterMove, err := taskRepo.ListByPlan(ctx, plan.ID)
		assert.NoError(t, err, "Failed to list tasks in first plan after move")
		assert.Equal(t, 0, len(tasksInFirstPlanAfterMove), "First plan should have 0 tasks after move")

		tasksInSecondPlan, err := taskRepo.ListByPlan(ctx, secondPlan.ID)
		assert.NoError(t, err, "Failed to list tasks in second plan")
		assert.Equal(t, 1, len(tasksInSecondPlan), "Second plan should have 1 task")
		assert.Equal(t, task.ID, tasksInSecondPlan[0].ID, "Task in second plan should match moved task")

		// Clean up
		err = taskRepo.Delete(ctx, task.ID)
		assert.NoError(t, err, "Failed to delete task")

		err = planRepo.Delete(ctx, secondPlan.ID)
		assert.NoError(t, err, "Failed to delete second plan")
	})
}

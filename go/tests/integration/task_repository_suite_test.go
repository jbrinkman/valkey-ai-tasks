package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/go/tests/utils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TaskRepositorySuite is a test suite for the TaskRepository
type TaskRepositorySuite struct {
	utils.RepositoryTestSuite
	TestPlan *models.Plan
}

// SetupTest sets up each test
func (s *TaskRepositorySuite) SetupTest() {
	// Call the base SetupTest to initialize the container and client
	s.RepositoryTestSuite.SetupTest()

	// Create a test plan for each test
	planRepo := s.GetPlanRepository()
	appID := "test-app-" + uuid.New().String()
	plan, err := planRepo.Create(s.Context, appID, "Test Plan", "Test plan description")
	require.NoError(s.T(), err, "Failed to create test plan")
	s.TestPlan = plan
}

// TestCreateTask tests creating a task
func (s *TaskRepositorySuite) TestCreateTask() {
	taskRepo := s.GetTaskRepository()

	task, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Test Task", "Test task description", models.TaskPriorityHigh)
	s.NoError(err, "Failed to create task")
	s.NotEmpty(task.ID, "Task ID should not be empty")
	s.Equal("Test Task", task.Title, "Task title should match")
	s.Equal("Test task description", task.Description, "Task description should match")
	s.Equal(models.TaskPriorityHigh, task.Priority, "Task priority should match")
	s.Equal(models.TaskStatusPending, task.Status, "Task should have default pending status")
	s.Equal(0, task.Order, "First task should have order 0")
	s.Equal(s.TestPlan.ID, task.PlanID, "Task should be associated with the correct plan")
}

// TestGetTask tests retrieving a task
func (s *TaskRepositorySuite) TestGetTask() {
	taskRepo := s.GetTaskRepository()

	// Create a task
	task, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Test Task", "Test task description", models.TaskPriorityHigh)
	s.NoError(err, "Failed to create task")

	// Get the task
	retrievedTask, err := taskRepo.Get(s.Context, task.ID)
	s.NoError(err, "Failed to get task")
	s.Equal(task.ID, retrievedTask.ID, "Task ID should match")
	s.Equal(task.Title, retrievedTask.Title, "Task title should match")
	s.Equal(task.Description, retrievedTask.Description, "Task description should match")
	s.Equal(task.Priority, retrievedTask.Priority, "Task priority should match")
	s.Equal(task.Status, retrievedTask.Status, "Task status should match")
	s.Equal(task.Order, retrievedTask.Order, "Task order should match")
	s.Equal(task.PlanID, retrievedTask.PlanID, "Task plan ID should match")
}

// TestGetNonExistentTask tests retrieving a non-existent task
func (s *TaskRepositorySuite) TestGetNonExistentTask() {
	taskRepo := s.GetTaskRepository()

	nonExistentID := uuid.New().String()
	_, err := taskRepo.Get(s.Context, nonExistentID)
	s.Error(err, "Getting non-existent task should return error")
	s.Contains(err.Error(), "task not found", "Error should indicate task not found")
}

// TestUpdateTask tests updating a task
func (s *TaskRepositorySuite) TestUpdateTask() {
	taskRepo := s.GetTaskRepository()

	// Create a task
	task, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Test Task", "Test task description", models.TaskPriorityHigh)
	s.NoError(err, "Failed to create task")

	// Update task properties
	task.Title = "Updated Task Title"
	task.Description = "Updated task description"
	task.Priority = models.TaskPriorityLow
	task.Status = models.TaskStatusInProgress

	// Perform the update
	err = taskRepo.Update(s.Context, task)
	s.NoError(err, "Failed to update task")

	// Retrieve the task again to verify updates
	updatedTask, err := taskRepo.Get(s.Context, task.ID)
	s.NoError(err, "Failed to get updated task")
	s.Equal("Updated Task Title", updatedTask.Title, "Task title should be updated")
	s.Equal("Updated task description", updatedTask.Description, "Task description should be updated")
	s.Equal(models.TaskPriorityLow, updatedTask.Priority, "Task priority should be updated")
	s.Equal(models.TaskStatusInProgress, updatedTask.Status, "Task status should be updated")
}

// TestUpdateNonExistentTask tests updating a non-existent task
func (s *TaskRepositorySuite) TestUpdateNonExistentTask() {
	taskRepo := s.GetTaskRepository()

	nonExistentTask := &models.Task{
		ID:          uuid.New().String(),
		Title:       "Non-existent Task",
		Description: "This task doesn't exist",
		PlanID:      s.TestPlan.ID,
	}
	err := taskRepo.Update(s.Context, nonExistentTask)
	s.Error(err, "Updating non-existent task should return error")
	s.Contains(err.Error(), "task not found", "Error should indicate task not found")
}

// TestListTasksByPlan tests listing tasks by plan
func (s *TaskRepositorySuite) TestListTasksByPlan() {
	taskRepo := s.GetTaskRepository()

	// Create tasks
	task1, err := taskRepo.Create(s.Context, s.TestPlan.ID, "First Task", "First task description", models.TaskPriorityHigh)
	s.NoError(err, "Failed to create first task")

	task2, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Second Task", "Second task description", models.TaskPriorityMedium)
	s.NoError(err, "Failed to create second task")

	task3, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Third Task", "Third task description", models.TaskPriorityLow)
	s.NoError(err, "Failed to create third task")

	// List tasks for the plan
	tasks, err := taskRepo.ListByPlan(s.Context, s.TestPlan.ID)
	s.NoError(err, "Failed to list tasks by plan")
	s.Equal(3, len(tasks), "Should have 3 tasks in the plan")

	// Verify tasks are ordered correctly
	s.Equal(0, tasks[0].Order, "First task should have order 0")
	s.Equal(1, tasks[1].Order, "Second task should have order 1")
	s.Equal(2, tasks[2].Order, "Third task should have order 2")

	// Verify task IDs match
	taskIDs := map[string]bool{
		task1.ID: false,
		task2.ID: false,
		task3.ID: false,
	}

	for _, task := range tasks {
		taskIDs[task.ID] = true
	}

	for id, found := range taskIDs {
		s.True(found, "Task with ID %s should be in the list", id)
	}
}

// TestListTasksByStatus tests listing tasks by status
func (s *TaskRepositorySuite) TestListTasksByStatus() {
	taskRepo := s.GetTaskRepository()

	// Create tasks with different statuses
	taskPending, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Pending Task", "A pending task", models.TaskPriorityMedium)
	s.NoError(err, "Failed to create pending task")

	taskInProgress, err := taskRepo.Create(s.Context, s.TestPlan.ID, "In Progress Task", "An in-progress task", models.TaskPriorityMedium)
	s.NoError(err, "Failed to create in-progress task")

	// Update the second task to in-progress status
	taskInProgress.Status = models.TaskStatusInProgress
	err = taskRepo.Update(s.Context, taskInProgress)
	s.NoError(err, "Failed to update task status")

	// List pending tasks
	pendingTasks, err := taskRepo.ListByStatus(s.Context, models.TaskStatusPending)
	s.NoError(err, "Failed to list pending tasks")

	// Find our specific pending task in the results
	foundPending := false
	for _, t := range pendingTasks {
		if t.ID == taskPending.ID {
			foundPending = true
			break
		}
	}
	s.True(foundPending, "Should find the pending task in pending tasks list")

	// List in-progress tasks
	inProgressTasks, err := taskRepo.ListByStatus(s.Context, models.TaskStatusInProgress)
	s.NoError(err, "Failed to list in-progress tasks")

	// Find our specific in-progress task in the results
	foundInProgress := false
	for _, t := range inProgressTasks {
		if t.ID == taskInProgress.ID {
			foundInProgress = true
			break
		}
	}
	s.True(foundInProgress, "Should find the in-progress task in in-progress tasks list")
}

// TestReorderTask tests reordering tasks
func (s *TaskRepositorySuite) TestReorderTask() {
	taskRepo := s.GetTaskRepository()

	// Create three tasks for reordering
	task1, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Task 1", "First task", models.TaskPriorityMedium)
	s.NoError(err, "Failed to create first task")

	task2, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Task 2", "Second task", models.TaskPriorityMedium)
	s.NoError(err, "Failed to create second task")

	task3, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Task 3", "Third task", models.TaskPriorityMedium)
	s.NoError(err, "Failed to create third task")

	// Verify initial order
	tasks, err := taskRepo.ListByPlan(s.Context, s.TestPlan.ID)
	s.NoError(err, "Failed to list tasks")
	s.Equal(3, len(tasks), "Should have 3 tasks")

	// Log initial task orders
	s.T().Logf("Initial orders - task1: %d, task2: %d, task3: %d",
		tasks[0].Order, tasks[1].Order, tasks[2].Order)

	// Move task1 to position 2 (last)
	err = taskRepo.ReorderTask(s.Context, task1.ID, 2)
	s.NoError(err, "Failed to reorder task")

	// Check the new order
	tasksAfterReorder, err := taskRepo.ListByPlan(s.Context, s.TestPlan.ID)
	s.NoError(err, "Failed to list tasks after reorder")

	// Verify the task order by checking the actual order in the returned list
	// The reordering should have moved task1 to the end, shifting task2 and task3 up
	s.Equal(3, len(tasksAfterReorder), "Should still have 3 tasks after reordering")

	// Get the order value for each task
	var task1Order, task2Order, task3Order int
	for _, task := range tasksAfterReorder {
		switch task.ID {
		case task1.ID:
			task1Order = task.Order
		case task2.ID:
			task2Order = task.Order
		case task3.ID:
			task3Order = task.Order
		}
	}

	// Log the actual orders after reordering
	s.T().Logf("After reorder - task1: %d, task2: %d, task3: %d",
		task1Order, task2Order, task3Order)

	// Update assertions to match the fixed implementation with 0-based ordering
	s.Equal(2, task1Order, "Task1 should now have order 2")
	s.Equal(0, task2Order, "Task2 should now have order 0")
	s.Equal(1, task3Order, "Task3 should now have order 1")
}

// TestDeleteTask tests deleting a task
func (s *TaskRepositorySuite) TestDeleteTask() {
	taskRepo := s.GetTaskRepository()

	// Create a task
	task, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Test Task", "Test task description", models.TaskPriorityHigh)
	s.NoError(err, "Failed to create task")

	// Delete the task
	err = taskRepo.Delete(s.Context, task.ID)
	s.NoError(err, "Failed to delete task")

	// Try to get the deleted task
	_, err = taskRepo.Get(s.Context, task.ID)
	s.Error(err, "Getting deleted task should return error")
	s.Contains(err.Error(), "task not found", "Error should indicate task not found")
}

// TestDeleteNonExistentTask tests deleting a non-existent task
func (s *TaskRepositorySuite) TestDeleteNonExistentTask() {
	taskRepo := s.GetTaskRepository()

	nonExistentID := uuid.New().String()
	err := taskRepo.Delete(s.Context, nonExistentID)
	s.Error(err, "Deleting non-existent task should return error")
	s.Contains(err.Error(), "task not found", "Error should indicate task not found")
}

// TestTaskRepositorySuite runs the task repository test suite
func TestTaskRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	suite.Run(t, new(TaskRepositorySuite))
}

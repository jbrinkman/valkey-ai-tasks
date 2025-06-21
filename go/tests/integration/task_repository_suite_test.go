package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/storage"
	"github.com/jbrinkman/valkey-ai-tasks/go/tests/utils"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TaskRepositorySuite is a test suite for the TaskRepository
// It includes both standard CRUD tests and edge case tests
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

// TestCreateTaskWithEmptyTitle tests creating a task with an empty title
func (s *TaskRepositorySuite) TestCreateTaskWithEmptyTitle() {
	taskRepo := s.GetTaskRepository()

	task, err := taskRepo.Create(s.Context, s.TestPlan.ID, "", "Task with empty title", models.TaskPriorityMedium)
	s.NoError(err, "Should be able to create task with empty title")
	s.Empty(task.Title, "Task title should be empty")
}

// TestCreateTaskWithEmptyDescription tests creating a task with an empty description
func (s *TaskRepositorySuite) TestCreateTaskWithEmptyDescription() {
	taskRepo := s.GetTaskRepository()

	task, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Task Title", "", models.TaskPriorityMedium)
	s.NoError(err, "Should be able to create task with empty description")
	s.Empty(task.Description, "Task description should be empty")
}

// TestCreateTaskWithNonExistentPlan tests creating a task with a non-existent plan
func (s *TaskRepositorySuite) TestCreateTaskWithNonExistentPlan() {
	taskRepo := s.GetTaskRepository()

	nonExistentPlanID := uuid.New().String()
	_, err := taskRepo.Create(s.Context, nonExistentPlanID, "Task Title", "Description", models.TaskPriorityMedium)
	s.Error(err, "Creating task with non-existent plan should fail")
	s.Contains(err.Error(), "plan not found", "Error should indicate plan not found")
}

// TestListTasksForNonExistentPlan tests listing tasks for a non-existent plan
func (s *TaskRepositorySuite) TestListTasksForNonExistentPlan() {
	taskRepo := s.GetTaskRepository()

	nonExistentPlanID := uuid.New().String()
	_, err := taskRepo.ListByPlan(s.Context, nonExistentPlanID)
	s.Error(err, "Listing tasks for non-existent plan should fail")
	s.Contains(err.Error(), "plan not found", "Error should indicate plan not found")
}

// TestReorderTaskWithInvalidOrder tests reordering a task with an invalid order
func (s *TaskRepositorySuite) TestReorderTaskWithInvalidOrder() {
	taskRepo := s.GetTaskRepository()

	// Create a task
	task, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Task for reordering", "Description", models.TaskPriorityMedium)
	s.NoError(err, "Failed to create task")

	// Try to reorder with invalid negative order
	err = taskRepo.ReorderTask(s.Context, task.ID, -1)
	s.Error(err, "Reordering task with negative order should fail")
	s.Contains(err.Error(), "invalid order", "Error should indicate invalid order")

	// Try to reorder with too large order
	err = taskRepo.ReorderTask(s.Context, task.ID, 100)
	s.Error(err, "Reordering task with too large order should fail")
	s.Contains(err.Error(), "invalid order", "Error should indicate invalid order")
}

// TestMoveTaskBetweenPlans tests moving a task between plans
func (s *TaskRepositorySuite) TestMoveTaskBetweenPlans() {
	taskRepo := s.GetTaskRepository()
	planRepo := s.GetPlanRepository()

	// Create a second plan
	appID := "test-app-" + uuid.New().String()
	secondPlan, err := planRepo.Create(s.Context, appID, "Second Plan", "Another plan")
	s.NoError(err, "Failed to create second plan")

	// Create a task in the first plan
	task, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Task to move", "This task will be moved", models.TaskPriorityMedium)
	s.NoError(err, "Failed to create task")

	// Verify task is in first plan
	tasksInFirstPlan, err := taskRepo.ListByPlan(s.Context, s.TestPlan.ID)
	s.NoError(err, "Failed to list tasks in first plan")
	s.Equal(1, len(tasksInFirstPlan), "First plan should have 1 task")

	// Move task to second plan
	taskToMove, err := taskRepo.Get(s.Context, task.ID)
	s.NoError(err, "Failed to get task")

	taskToMove.PlanID = secondPlan.ID
	err = taskRepo.Update(s.Context, taskToMove)
	s.NoError(err, "Failed to update task's plan")

	// Verify task is now in second plan
	tasksInFirstPlanAfterMove, err := taskRepo.ListByPlan(s.Context, s.TestPlan.ID)
	s.NoError(err, "Failed to list tasks in first plan after move")
	s.Equal(0, len(tasksInFirstPlanAfterMove), "First plan should have 0 tasks after move")

	tasksInSecondPlan, err := taskRepo.ListByPlan(s.Context, secondPlan.ID)
	s.NoError(err, "Failed to list tasks in second plan")
	s.Equal(1, len(tasksInSecondPlan), "Second plan should have 1 task")
	s.Equal(task.ID, tasksInSecondPlan[0].ID, "Task in second plan should match moved task")
}

// TestCreateTaskWithSpecialCharacters tests creating a task with special characters
func (s *TaskRepositorySuite) TestCreateTaskWithSpecialCharacters() {
	taskRepo := s.GetTaskRepository()

	specialTitle := "Special!@#$%^&*()_+{}[]|\\:;\"'<>,.?/~` Task"
	specialDesc := "Description with emoji 😀 and unicode characters: ñáéíóú"

	task, err := taskRepo.Create(s.Context, s.TestPlan.ID, specialTitle, specialDesc, models.TaskPriorityMedium)
	s.NoError(err, "Should be able to create task with special characters")

	// Retrieve the task to verify special characters are preserved
	retrievedTask, err := taskRepo.Get(s.Context, task.ID)
	s.NoError(err, "Failed to get task with special characters")
	s.Equal(specialTitle, retrievedTask.Title, "Special characters in title should be preserved")
	s.Equal(specialDesc, retrievedTask.Description, "Special characters in description should be preserved")
}

// TestBulkTaskOperations tests bulk operations on tasks
func (s *TaskRepositorySuite) TestBulkTaskOperations() {
	taskRepo := s.GetTaskRepository()

	// Create multiple tasks
	taskInputs := []struct {
		title       string
		description string
		priority    models.TaskPriority
	}{
		{"Task 1", "Description 1", models.TaskPriorityHigh},
		{"Task 2", "Description 2", models.TaskPriorityMedium},
		{"Task 3", "Description 3", models.TaskPriorityLow},
	}

	var tasks []*models.Task
	for _, input := range taskInputs {
		task, err := taskRepo.Create(s.Context, s.TestPlan.ID, input.title, input.description, input.priority)
		s.NoError(err, "Failed to create task")
		tasks = append(tasks, task)
	}

	// Update all tasks to completed status
	for _, task := range tasks {
		task.Status = models.TaskStatusCompleted
		err := taskRepo.Update(s.Context, task)
		s.NoError(err, "Failed to update task status")
	}

	// List completed tasks
	completedTasks, err := taskRepo.ListByStatus(s.Context, models.TaskStatusCompleted)
	s.NoError(err, "Failed to list completed tasks")

	// Verify our tasks are in the completed list
	taskIDs := make(map[string]bool)
	for _, task := range tasks {
		taskIDs[task.ID] = false
	}

	for _, task := range completedTasks {
		if _, exists := taskIDs[task.ID]; exists {
			taskIDs[task.ID] = true
		}
	}

	for id, found := range taskIDs {
		s.True(found, "Task %s should be in completed tasks list", id)
	}

	// Delete all tasks
	for _, task := range tasks {
		err := taskRepo.Delete(s.Context, task.ID)
		s.NoError(err, "Failed to delete task")
	}

	// Verify plan has no tasks
	remainingTasks, err := taskRepo.ListByPlan(s.Context, s.TestPlan.ID)
	s.NoError(err, "Failed to list remaining tasks")
	s.Empty(remainingTasks, "Plan should have no tasks after deletion")
}

// TestCreateBulkTasks tests the bulk task creation functionality
func (s *TaskRepositorySuite) TestCreateBulkTasks() {
	taskRepo := s.GetTaskRepository()

	// Prepare task inputs for bulk creation
	taskInputs := []storage.TaskCreateInput{
		{
			Title:       "Task 1",
			Description: "Description for task 1",
			Priority:    models.TaskPriorityHigh,
		},
		{
			Title:       "Task 2",
			Description: "Description for task 2",
			Status:      models.TaskStatusInProgress,
		},
		{
			Title: "Task 3",
		},
	}

	// Create tasks in bulk
	createdTasks, err := taskRepo.CreateBulk(s.Context, s.TestPlan.ID, taskInputs)
	s.NoError(err, "Failed to create tasks in bulk")
	s.NotNil(createdTasks, "Created tasks should not be nil")
	s.Equal(3, len(createdTasks), "Should have created 3 tasks")

	// Verify task 1
	s.Equal("Task 1", createdTasks[0].Title)
	s.Equal("Description for task 1", createdTasks[0].Description)
	s.Equal(models.TaskPriorityHigh, createdTasks[0].Priority)
	s.Equal(models.TaskStatusPending, createdTasks[0].Status) // Default status
	s.Equal(0, createdTasks[0].Order)

	// Verify task 2
	s.Equal("Task 2", createdTasks[1].Title)
	s.Equal("Description for task 2", createdTasks[1].Description)
	s.Equal(models.TaskPriorityMedium, createdTasks[1].Priority) // Default priority
	s.Equal(models.TaskStatusInProgress, createdTasks[1].Status)
	s.Equal(1, createdTasks[1].Order)

	// Verify task 3
	s.Equal("Task 3", createdTasks[2].Title)
	s.Equal("no description provided", createdTasks[2].Description) // Default description
	s.Equal(models.TaskPriorityMedium, createdTasks[2].Priority)    // Default priority
	s.Equal(models.TaskStatusPending, createdTasks[2].Status)       // Default status
	s.Equal(2, createdTasks[2].Order)

	// Verify tasks are stored in Valkey
	tasks, err := taskRepo.ListByPlan(s.Context, s.TestPlan.ID)
	s.NoError(err, "Failed to list tasks by plan")
	s.Equal(3, len(tasks), "Should have 3 tasks in the plan")
}

// TestListTasksByPlanAndStatus tests listing tasks by both plan ID and status
func (s *TaskRepositorySuite) TestListTasksByPlanAndStatus() {
	taskRepo := s.GetTaskRepository()

	// Create a second plan to ensure filtering by plan works
	planRepo := s.GetPlanRepository()
	appID := "test-app-" + uuid.New().String()
	secondPlan, err := planRepo.Create(s.Context, appID, "Second Plan", "Second plan description")
	s.NoError(err, "Failed to create second test plan")

	// Create tasks with different statuses in both plans
	// Tasks in the first plan
	task1, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Task 1", "Description 1", models.TaskPriorityMedium)
	s.NoError(err, "Failed to create task 1")
	task1.Status = models.TaskStatusInProgress
	err = taskRepo.Update(s.Context, task1)
	s.NoError(err, "Failed to update task 1 status")

	task2, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Task 2", "Description 2", models.TaskPriorityMedium)
	s.NoError(err, "Failed to create task 2")
	task2.Status = models.TaskStatusCompleted
	err = taskRepo.Update(s.Context, task2)
	s.NoError(err, "Failed to update task 2 status")

	task3, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Task 3", "Description 3", models.TaskPriorityMedium)
	s.NoError(err, "Failed to create task 3")
	task3.Status = models.TaskStatusInProgress
	err = taskRepo.Update(s.Context, task3)
	s.NoError(err, "Failed to update task 3 status")

	// Tasks in the second plan
	task4, err := taskRepo.Create(s.Context, secondPlan.ID, "Task 4", "Description 4", models.TaskPriorityMedium)
	s.NoError(err, "Failed to create task 4")
	task4.Status = models.TaskStatusInProgress
	err = taskRepo.Update(s.Context, task4)
	s.NoError(err, "Failed to update task 4 status")

	task5, err := taskRepo.Create(s.Context, secondPlan.ID, "Task 5", "Description 5", models.TaskPriorityMedium)
	s.NoError(err, "Failed to create task 5")
	task5.Status = models.TaskStatusPending
	err = taskRepo.Update(s.Context, task5)
	s.NoError(err, "Failed to update task 5 status")

	// Test filtering by plan and status
	// Get in-progress tasks from the first plan
	inProgressTasks, err := taskRepo.ListByPlanAndStatus(s.Context, s.TestPlan.ID, models.TaskStatusInProgress)
	s.NoError(err, "Failed to list tasks by plan and status")
	s.Equal(2, len(inProgressTasks), "Should find 2 in-progress tasks in the first plan")
	s.Equal(task1.ID, inProgressTasks[0].ID, "First task should match")
	s.Equal(task3.ID, inProgressTasks[1].ID, "Second task should match")

	// Get completed tasks from the first plan
	completedTasks, err := taskRepo.ListByPlanAndStatus(s.Context, s.TestPlan.ID, models.TaskStatusCompleted)
	s.NoError(err, "Failed to list tasks by plan and status")
	s.Equal(1, len(completedTasks), "Should find 1 completed task in the first plan")
	s.Equal(task2.ID, completedTasks[0].ID, "Completed task should match")

	// Get pending tasks from the first plan (should be empty)
	pendingTasks, err := taskRepo.ListByPlanAndStatus(s.Context, s.TestPlan.ID, models.TaskStatusPending)
	s.NoError(err, "Failed to list tasks by plan and status")
	s.Equal(0, len(pendingTasks), "Should find 0 pending tasks in the first plan")

	// Get in-progress tasks from the second plan
	secondPlanInProgressTasks, err := taskRepo.ListByPlanAndStatus(s.Context, secondPlan.ID, models.TaskStatusInProgress)
	s.NoError(err, "Failed to list tasks by plan and status")
	s.Equal(1, len(secondPlanInProgressTasks), "Should find 1 in-progress task in the second plan")
	s.Equal(task4.ID, secondPlanInProgressTasks[0].ID, "In-progress task in second plan should match")

	// Get pending tasks from the second plan
	secondPlanPendingTasks, err := taskRepo.ListByPlanAndStatus(s.Context, secondPlan.ID, models.TaskStatusPending)
	s.NoError(err, "Failed to list tasks by plan and status")
	s.Equal(1, len(secondPlanPendingTasks), "Should find 1 pending task in the second plan")
	s.Equal(task5.ID, secondPlanPendingTasks[0].ID, "Pending task in second plan should match")

	// Test with non-existent plan
	nonExistentPlanTasks, err := taskRepo.ListByPlanAndStatus(s.Context, "non-existent-plan", models.TaskStatusInProgress)
	s.Error(err, "Should error when listing tasks for non-existent plan")
	s.Nil(nonExistentPlanTasks, "Should return nil tasks for non-existent plan")
}

// TestMCPBulkCreateTasks tests the MCP bulk_create_tasks tool
func (s *TaskRepositorySuite) TestMCPBulkCreateTasks() {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only handle POST requests to /call-tool/bulk_create_tasks
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/call-tool/bulk_create_tasks") {
			// Parse the request body
			var requestBody struct {
				PlanID    string `json:"plan_id"`
				TasksJSON string `json:"tasks_json"`
			}
			err := json.NewDecoder(r.Body).Decode(&requestBody)
			if err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// Validate required fields
			if requestBody.PlanID == "" || requestBody.TasksJSON == "" {
				http.Error(w, "Missing required fields", http.StatusBadRequest)
				return
			}

			// Parse the tasks JSON
			var taskInputs []map[string]interface{}
			err = json.Unmarshal([]byte(requestBody.TasksJSON), &taskInputs)
			if err != nil {
				http.Error(w, "Invalid tasks JSON", http.StatusBadRequest)
				return
			}

			// Convert to TaskCreateInput
			var inputs []storage.TaskCreateInput
			for _, task := range taskInputs {
				input := storage.TaskCreateInput{
					Title:       task["title"].(string),
					Description: "",
					Status:      models.TaskStatusPending,
					Priority:    models.TaskPriorityMedium,
				}

				if desc, ok := task["description"].(string); ok {
					input.Description = desc
				}

				if status, ok := task["status"].(string); ok {
					input.Status = models.TaskStatus(status)
				}

				if priority, ok := task["priority"].(string); ok {
					input.Priority = models.TaskPriority(priority)
				}

				inputs = append(inputs, input)
			}

			// Create tasks in bulk
			taskRepo := s.GetTaskRepository()
			createdTasks, err := taskRepo.CreateBulk(s.Context, requestBody.PlanID, inputs)
			if err != nil {
				http.Error(w, "Failed to create tasks: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Return the created tasks
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(createdTasks)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create a test HTTP client
	client := server.Client()

	// Prepare the request body
	taskInputs := []map[string]interface{}{
		{
			"title":       "Task 1",
			"description": "Description for task 1",
			"priority":    "high",
		},
		{
			"title":       "Task 2",
			"description": "Description for task 2",
			"status":      "in_progress",
		},
		{
			"title": "Task 3",
		},
	}
	taskInputsJSON, err := json.Marshal(taskInputs)
	s.Require().NoError(err, "Failed to marshal task inputs")

	requestBody := map[string]string{
		"plan_id":    s.TestPlan.ID,
		"tasks_json": string(taskInputsJSON),
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	s.Require().NoError(err, "Failed to marshal request body")

	// Send the request
	resp, err := client.Post(server.URL+"/call-tool/bulk_create_tasks", "application/json", strings.NewReader(string(requestBodyJSON)))
	s.Require().NoError(err, "Failed to send request")
	defer resp.Body.Close()

	// Check the response status code
	s.Equal(http.StatusOK, resp.StatusCode, "Expected status code 200")

	// Parse the response body
	var createdTasks []models.Task
	err = json.NewDecoder(resp.Body).Decode(&createdTasks)
	s.Require().NoError(err, "Failed to parse response body")

	// Verify the created tasks
	s.Equal(3, len(createdTasks), "Should have created 3 tasks")

	// Verify task 1
	s.Equal("Task 1", createdTasks[0].Title)
	s.Equal("Description for task 1", createdTasks[0].Description)
	s.Equal(models.TaskPriorityHigh, createdTasks[0].Priority)
	s.Equal(models.TaskStatusPending, createdTasks[0].Status) // Default status
	s.Equal(0, createdTasks[0].Order)

	// Verify task 2
	s.Equal("Task 2", createdTasks[1].Title)
	s.Equal("Description for task 2", createdTasks[1].Description)
	s.Equal(models.TaskPriorityMedium, createdTasks[1].Priority) // Default priority
	s.Equal(models.TaskStatusInProgress, createdTasks[1].Status)
	s.Equal(1, createdTasks[1].Order)

	// Verify task 3
	s.Equal("Task 3", createdTasks[2].Title)
	s.Equal("no description provided", createdTasks[2].Description) // Default description
	s.Equal(models.TaskPriorityMedium, createdTasks[2].Priority)    // Default priority
	s.Equal(models.TaskStatusPending, createdTasks[2].Status)       // Default status
	s.Equal(2, createdTasks[2].Order)

	// Verify tasks are stored in Valkey
	taskRepo := s.GetTaskRepository()
	tasks, err := taskRepo.ListByPlan(s.Context, s.TestPlan.ID)
	s.Require().NoError(err, "Failed to list tasks by plan")
	s.Equal(3, len(tasks), "Should have 3 tasks in the plan")
}

// TestTaskRepositorySuite runs the task repository test suite
func TestTaskRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	suite.Run(t, new(TaskRepositorySuite))
}

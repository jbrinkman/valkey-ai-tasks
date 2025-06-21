package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/go/tests/utils"
	"github.com/stretchr/testify/suite"
)

// TaskRepositoryEdgeSuite is a test suite for the TaskRepository edge cases
type TaskRepositoryEdgeSuite struct {
	utils.RepositoryTestSuite
	TestPlan *models.Plan
}

// SetupTest sets up each test
func (s *TaskRepositoryEdgeSuite) SetupTest() {
	// Call the base SetupTest to initialize the container and client
	s.RepositoryTestSuite.SetupTest()

	// Create a test plan for each test
	planRepo := s.GetPlanRepository()
	appID := "test-app-" + uuid.New().String()
	plan, err := planRepo.Create(s.Context, appID, "Test Plan", "Test plan description")
	s.Require().NoError(err, "Failed to create test plan")
	s.TestPlan = plan
}

// TestCreateTaskWithEmptyTitle tests creating a task with an empty title
func (s *TaskRepositoryEdgeSuite) TestCreateTaskWithEmptyTitle() {
	taskRepo := s.GetTaskRepository()

	task, err := taskRepo.Create(s.Context, s.TestPlan.ID, "", "Task with empty title", models.TaskPriorityMedium)
	s.NoError(err, "Should be able to create task with empty title")
	s.Empty(task.Title, "Task title should be empty")
}

// TestCreateTaskWithEmptyDescription tests creating a task with an empty description
func (s *TaskRepositoryEdgeSuite) TestCreateTaskWithEmptyDescription() {
	taskRepo := s.GetTaskRepository()

	task, err := taskRepo.Create(s.Context, s.TestPlan.ID, "Task Title", "", models.TaskPriorityMedium)
	s.NoError(err, "Should be able to create task with empty description")
	s.Empty(task.Description, "Task description should be empty")
}

// TestCreateTaskWithNonExistentPlan tests creating a task with a non-existent plan
func (s *TaskRepositoryEdgeSuite) TestCreateTaskWithNonExistentPlan() {
	taskRepo := s.GetTaskRepository()

	nonExistentPlanID := uuid.New().String()
	_, err := taskRepo.Create(s.Context, nonExistentPlanID, "Task Title", "Description", models.TaskPriorityMedium)
	s.Error(err, "Creating task with non-existent plan should fail")
	s.Contains(err.Error(), "plan not found", "Error should indicate plan not found")
}

// TestListTasksForNonExistentPlan tests listing tasks for a non-existent plan
func (s *TaskRepositoryEdgeSuite) TestListTasksForNonExistentPlan() {
	taskRepo := s.GetTaskRepository()

	nonExistentPlanID := uuid.New().String()
	_, err := taskRepo.ListByPlan(s.Context, nonExistentPlanID)
	s.Error(err, "Listing tasks for non-existent plan should fail")
	s.Contains(err.Error(), "plan not found", "Error should indicate plan not found")
}

// TestReorderTaskWithInvalidOrder tests reordering a task with an invalid order
func (s *TaskRepositoryEdgeSuite) TestReorderTaskWithInvalidOrder() {
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
func (s *TaskRepositoryEdgeSuite) TestMoveTaskBetweenPlans() {
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
func (s *TaskRepositoryEdgeSuite) TestCreateTaskWithSpecialCharacters() {
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
func (s *TaskRepositoryEdgeSuite) TestBulkTaskOperations() {
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

// TestTaskRepositoryEdgeSuite runs the task repository edge case test suite
func TestTaskRepositoryEdgeSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	suite.Run(t, new(TaskRepositoryEdgeSuite))
}

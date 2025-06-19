package integration

import (
	"context"
	"testing"

	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/storage"
	"github.com/jbrinkman/valkey-ai-tasks/go/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskRepositoryCreateBulk tests the bulk task creation functionality
func TestTaskRepositoryCreateBulk(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set up test context
	ctx := context.Background()
	req := require.New(t)

	// Start a Valkey container
	container, err := utils.StartValkeyContainer(ctx, t)
	req.NoError(err, "Failed to start Valkey container")
	defer utils.StopValkeyContainer(ctx, t, container)

	// Extract host and port from container URI
	// URI format is redis://host:port
	uri := container.URI
	host := "localhost" // Default to localhost
	port := "6379"      // Default port

	// Parse the URI to extract host and port if needed
	if len(uri) > 8 { // "redis://" is 8 chars
		hostPort := uri[8:] // Remove "redis://" prefix
		if hostPort != "" {
			parts := utils.ParseHostPort(hostPort)
			if len(parts) == 2 {
				host = parts[0]
				port = parts[1]
			}
		}
	}

	// Create Valkey client
	valkeyClient, err := storage.NewValkeyClient(host, utils.ParseInt(port), "", "")
	req.NoError(err, "Failed to create Valkey client")
	defer valkeyClient.Close()

	// Create repositories
	projectRepo := storage.NewProjectRepository(valkeyClient)
	taskRepo := storage.NewTaskRepository(valkeyClient)

	// Create a test project
	project, err := projectRepo.Create(ctx, "test-app", "Test Project", "Test project description")
	req.NoError(err, "Failed to create test project")

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
	createdTasks, err := taskRepo.CreateBulk(ctx, project.ID, taskInputs)
	req.NoError(err, "Failed to create tasks in bulk")
	req.NotNil(createdTasks, "Created tasks should not be nil")
	req.Equal(3, len(createdTasks), "Should have created 3 tasks")

	// Verify the created tasks
	assert := assert.New(t)

	// Verify task 1
	assert.Equal("Task 1", createdTasks[0].Title)
	assert.Equal("Description for task 1", createdTasks[0].Description)
	assert.Equal(models.TaskPriorityHigh, createdTasks[0].Priority)
	assert.Equal(models.TaskStatusPending, createdTasks[0].Status) // Default status
	assert.Equal(0, createdTasks[0].Order)

	// Verify task 2
	assert.Equal("Task 2", createdTasks[1].Title)
	assert.Equal("Description for task 2", createdTasks[1].Description)
	assert.Equal(models.TaskPriorityMedium, createdTasks[1].Priority) // Default priority
	assert.Equal(models.TaskStatusInProgress, createdTasks[1].Status)
	assert.Equal(1, createdTasks[1].Order)

	// Verify task 3
	assert.Equal("Task 3", createdTasks[2].Title)
	assert.Equal("no description provided", createdTasks[2].Description) // Default description
	assert.Equal(models.TaskPriorityMedium, createdTasks[2].Priority)    // Default priority
	assert.Equal(models.TaskStatusPending, createdTasks[2].Status)       // Default status
	assert.Equal(2, createdTasks[2].Order)

	// Verify tasks are stored in Valkey
	tasks, err := taskRepo.ListByProject(ctx, project.ID)
	req.NoError(err, "Failed to list tasks by project")
	assert.Equal(3, len(tasks), "Should have 3 tasks in the project")
}

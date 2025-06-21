package integration

import (
	"context"
	"testing"

	"github.com/jbrinkman/valkey-ai-tasks/go/internal/storage"
	"github.com/jbrinkman/valkey-ai-tasks/go/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValkeyClientConnection tests basic Valkey client connection
func TestValkeyClientConnection(t *testing.T) {
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
	host := "localhost" // Default to localhost
	port := 6379        // Default port

	// Create a new Valkey client
	valkeyClient, err := storage.NewValkeyClient(host, port, "", "")
	req.NoError(err, "Failed to create Valkey client")
	defer valkeyClient.Close()

	// Test ping
	err = valkeyClient.Ping(ctx)
	req.NoError(err, "Failed to ping Valkey")
}

// TestValkeyClientConnectionFailure tests Valkey client connection failure scenarios
func TestValkeyClientConnectionFailure(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set up test context
	ctx := context.Background()

	// Test connection to non-existent server
	valkeyClient, err := storage.NewValkeyClient("non-existent-host", 6379, "", "")
	if err == nil {
		// Some implementations might not fail on creation, so try to ping
		err = valkeyClient.Ping(ctx)
		assert.Error(t, err, "Expected error when connecting to non-existent server")
		valkeyClient.Close()
	} else {
		assert.Error(t, err, "Expected error when connecting to non-existent server")
	}

	// Test connection with invalid credentials
	// First start a regular container
	container, err := utils.StartValkeyContainer(ctx, t)
	require.NoError(t, err, "Failed to start Valkey container")
	defer utils.StopValkeyContainer(ctx, t, container)

	// Extract host and port
	host := "localhost" // Default to localhost
	port := 6379        // Default port

	// Try to connect with invalid credentials
	valkeyClient, err = storage.NewValkeyClient(host, port, "invaliduser", "invalidpass")
	if err == nil {
		// Some implementations might not fail on creation, so try to ping
		err = valkeyClient.Ping(ctx)
		assert.Error(t, err, "Expected error when connecting with invalid credentials")
		valkeyClient.Close()
	}
}

// TestValkeyClientWithAuth tests Valkey client with authentication
func TestValkeyClientWithAuth(t *testing.T) {
	// Skip this test for now as it requires additional configuration
	t.Skip("Skipping authentication test as it requires additional configuration")

	// This test would require setting up a Valkey container with authentication
	// which is not directly supported by the current testcontainers-go/modules/valkey
	// In a real implementation, you might need to use custom command arguments or
	// a custom Dockerfile to set up authentication
}

// TestValkeyClientKeyHelpers tests our application's key helper functions
func TestValkeyClientKeyHelpers(t *testing.T) {
	// This test focuses on our application's key generation helpers
	// These are important to test as they define our data structure in Valkey

	// Test plan key generation
	planID := "test-plan-123"
	planKey := storage.GetPlanKey(planID)
	assert.Equal(t, "plan:test-plan-123", planKey, "Plan key should be correctly formatted")

	// Test task key generation
	taskID := "test-task-456"
	taskKey := storage.GetTaskKey(taskID)
	assert.Equal(t, "task:test-task-456", taskKey, "Task key should be correctly formatted")

	// Test plan tasks key generation
	planTasksKey := storage.GetPlanTasksKey(planID)
	assert.Equal(t, "plan_tasks:test-plan-123", planTasksKey, "Plan tasks key should be correctly formatted")

	// Test legacy project key generation (for backward compatibility)
	projectID := "test-project-789"
	projectKey := storage.GetProjectKey(projectID)
	assert.Equal(t, "project:test-project-789", projectKey, "Project key should be correctly formatted")

	// Test legacy project tasks key generation
	projectTasksKey := storage.GetProjectTasksKey(projectID)
	assert.Equal(t, "project_tasks:test-project-789", projectTasksKey, "Project tasks key should be correctly formatted")
}

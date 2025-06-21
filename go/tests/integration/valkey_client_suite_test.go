package integration

import (
	"context"
	"testing"

	"github.com/jbrinkman/valkey-ai-tasks/go/internal/storage"
	"github.com/jbrinkman/valkey-ai-tasks/go/tests/utils"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
)

// ValkeyClientSuite is a test suite for the Valkey client
type ValkeyClientSuite struct {
	suite.Suite
	Context   context.Context
	Container testcontainers.Container
}

// SetupSuite prepares the test suite
func (s *ValkeyClientSuite) SetupSuite() {
	// Skip in short mode
	if testing.Short() {
		s.T().Skip("Skipping integration test in short mode")
	}

	// Set up test context
	s.Context = context.Background()
}

// TearDownSuite cleans up after the test suite
func (s *ValkeyClientSuite) TearDownSuite() {
	// Nothing to clean up at the suite level
}

// TestConnection tests basic Valkey client connection
func (s *ValkeyClientSuite) TestConnection() {
	// Start a Valkey container
	container, err := utils.StartValkeyContainer(s.Context, s.T())
	s.Require().NoError(err, "Failed to start Valkey container")
	defer utils.StopValkeyContainer(s.Context, s.T(), container)

	// Extract host and port from container URI
	host := "localhost" // Default to localhost
	port := 6379        // Default port

	// Create a new Valkey client
	valkeyClient, err := storage.NewValkeyClient(host, port, "", "")
	s.Require().NoError(err, "Failed to create Valkey client")
	defer valkeyClient.Close()

	// Test ping
	err = valkeyClient.Ping(s.Context)
	s.Require().NoError(err, "Failed to ping Valkey")
}

// TestConnectionFailure tests Valkey client connection failure scenarios
func (s *ValkeyClientSuite) TestConnectionFailure() {
	// Test connection to non-existent server
	valkeyClient, err := storage.NewValkeyClient("non-existent-host", 6379, "", "")
	if err == nil {
		// Some implementations might not fail on creation, so try to ping
		err = valkeyClient.Ping(s.Context)
		s.Assert().Error(err, "Expected error when connecting to non-existent server")
		valkeyClient.Close()
	} else {
		s.Assert().Error(err, "Expected error when connecting to non-existent server")
	}

	// Test connection with invalid credentials
	// First start a regular container
	container, err := utils.StartValkeyContainer(s.Context, s.T())
	s.Require().NoError(err, "Failed to start Valkey container")
	defer utils.StopValkeyContainer(s.Context, s.T(), container)

	// Extract host and port
	host := "localhost" // Default to localhost
	port := 6379        // Default port

	// Try to connect with invalid credentials
	valkeyClient, err = storage.NewValkeyClient(host, port, "invaliduser", "invalidpass")
	if err == nil {
		// Some implementations might not fail on creation, so try to ping
		err = valkeyClient.Ping(s.Context)
		s.Assert().Error(err, "Expected error when connecting with invalid credentials")
		valkeyClient.Close()
	}
}

// TestWithAuth tests Valkey client with authentication
func (s *ValkeyClientSuite) TestWithAuth() {
	// Skip this test for now as it requires additional configuration
	s.T().Skip("Skipping authentication test as it requires additional configuration")

	// This test would require setting up a Valkey container with authentication
	// which is not directly supported by the current testcontainers-go/modules/valkey
	// In a real implementation, you might need to use custom command arguments or
	// a custom Dockerfile to set up authentication
}

// TestKeyHelpers tests our application's key helper functions
func (s *ValkeyClientSuite) TestKeyHelpers() {
	// This test focuses on our application's key generation helpers
	// These are important to test as they define our data structure in Valkey

	// Test plan key generation
	planID := "test-plan-123"
	planKey := storage.GetPlanKey(planID)
	s.Assert().Equal("plan:test-plan-123", planKey, "Plan key should be correctly formatted")

	// Test task key generation
	taskID := "test-task-456"
	taskKey := storage.GetTaskKey(taskID)
	s.Assert().Equal("task:test-task-456", taskKey, "Task key should be correctly formatted")

	// Test plan tasks key generation
	planTasksKey := storage.GetPlanTasksKey(planID)
	s.Assert().Equal("plan_tasks:test-plan-123", planTasksKey, "Plan tasks key should be correctly formatted")

	// Test legacy project key generation (for backward compatibility)
	projectID := "test-project-789"
	projectKey := storage.GetProjectKey(projectID)
	s.Assert().Equal("project:test-project-789", projectKey, "Project key should be correctly formatted")

	// Test legacy project tasks key generation
	projectTasksKey := storage.GetProjectTasksKey(projectID)
	s.Assert().Equal("project_tasks:test-project-789", projectTasksKey, "Project tasks key should be correctly formatted")
}

// TestValkeyClientSuite runs the Valkey client test suite
func TestValkeyClientSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	suite.Run(t, new(ValkeyClientSuite))
}

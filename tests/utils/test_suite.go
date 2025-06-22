package utils

import (
	"context"
	"testing"

	"github.com/jbrinkman/valkey-ai-tasks/internal/storage"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RepositoryTestSuite is a base test suite for repository tests
type RepositoryTestSuite struct {
	suite.Suite
	Context      context.Context
	Containers   []*ValkeyContainer
	ValkeyClient *storage.ValkeyClient
}

// SetupSuite sets up the test suite
func (s *RepositoryTestSuite) SetupSuite() {
	// Create context
	s.Context = context.Background()
}

// TearDownSuite tears down the test suite
func (s *RepositoryTestSuite) TearDownSuite() {
	// Close Valkey client if it exists
	if s.ValkeyClient != nil {
		s.ValkeyClient.Close()
	}

	// Stop all containers
	for _, container := range s.Containers {
		StopValkeyContainer(s.Context, s.T(), container)
	}
}

// SetupTest sets up each test
func (s *RepositoryTestSuite) SetupTest() {
	// Skip in short mode
	if testing.Short() {
		s.T().Skip("Skipping integration test in short mode")
	}

	// Start a Valkey container
	container, err := StartValkeyContainer(s.Context, s.T())
	require.NoError(s.T(), err, "Failed to start Valkey container")

	// Add container to list for cleanup
	s.Containers = append(s.Containers, container)

	// Create Valkey client using the container's endpoint
	endpoint, err := container.Container.Endpoint(s.Context, "")
	require.NoError(s.T(), err, "Failed to get container endpoint")

	// Parse the endpoint to get host and port
	host, port, err := ParseEndpoint(endpoint)
	require.NoError(s.T(), err, "Failed to parse container endpoint")

	valkeyClient, err := storage.NewValkeyClient(host, port, "", "")
	require.NoError(s.T(), err, "Failed to create Valkey client")
	s.ValkeyClient = valkeyClient
}

// TearDownTest cleans up after each test
func (s *RepositoryTestSuite) TearDownTest() {
	// Flush all data in the current Valkey container
	if s.ValkeyClient != nil && len(s.Containers) > 0 {
		// Use the container's client to flush all data
		container := s.Containers[len(s.Containers)-1]
		if container != nil && container.Client != nil {
			_, err := container.Client.FlushAll(s.Context)
			s.NoError(err, "Failed to flush Valkey database")
		}
	}
}

// GetPlanRepository returns a new plan repository
func (s *RepositoryTestSuite) GetPlanRepository() *storage.PlanRepository {
	return storage.NewPlanRepository(s.ValkeyClient)
}

// GetTaskRepository returns a new task repository
func (s *RepositoryTestSuite) GetTaskRepository() *storage.TaskRepository {
	return storage.NewTaskRepository(s.ValkeyClient)
}

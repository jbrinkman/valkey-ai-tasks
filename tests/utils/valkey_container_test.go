package utils_test

import (
	"testing"

	"github.com/jbrinkman/valkey-ai-tasks/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValkeyContainer tests the Valkey container utilities
func TestValkeyContainer(t *testing.T) {
	// Skip this test in CI environments or when Docker is not available
	// This is an integration test that requires Docker
	if testing.Short() {
		t.Skip("Skipping Valkey container test in short mode")
	}

	// Set up Valkey container
	ctx, container, cleanup := utils.SetupValkeyTest(t)
	defer cleanup()

	// Verify that the container is running and accessible
	assert.NotNil(t, container)
	assert.NotNil(t, container.Client)
	assert.NotEmpty(t, container.URI)

	// Test connection to Valkey
	pong, err := container.Client.Ping(ctx)
	require.NoError(t, err, "Failed to ping Valkey container")
	assert.Equal(t, "PONG", pong)

	// Test basic operations
	result, err := container.Client.Set(ctx, "test:key", "test-value")
	require.NoError(t, err, "Failed to set value in Valkey")
	assert.Equal(t, "OK", result)

	resultGet, err := container.Client.Get(ctx, "test:key")
	require.NoError(t, err, "Failed to get value from Valkey")
	assert.Equal(t, "test-value", resultGet.Value())
}

// TestPopulateTestData tests the test data generation utilities
func TestPopulateTestData(t *testing.T) {
	// Skip this test in CI environments or when Docker is not available
	if testing.Short() {
		t.Skip("Skipping test data population test in short mode")
	}

	// Set up Valkey container
	ctx, container, cleanup := utils.SetupValkeyTest(t)
	defer cleanup()

	// Populate test data
	projectCount := 2
	tasksPerProject := 3
	projectIDs, taskIDs := utils.PopulateTestData(ctx, t, container.Client, projectCount, tasksPerProject)

	// Verify project count
	assert.Len(t, projectIDs, projectCount)

	// Verify task count
	assert.Len(t, taskIDs, projectCount*tasksPerProject)

	// Verify projects exist in Valkey
	for _, projectID := range projectIDs {
		exists, err := container.Client.Exists(ctx, []string{"project:" + projectID})
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)
	}

	// Verify tasks exist in Valkey
	for _, taskID := range taskIDs {
		exists, err := container.Client.Exists(ctx, []string{"task:" + taskID})
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)
	}

	// Test cleanup
	utils.CleanupValkeyData(ctx, t, container.Client)

	// Verify data was cleaned up
	for _, projectID := range projectIDs {
		exists, err := container.Client.Exists(ctx, []string{"project:" + projectID})
		require.NoError(t, err)
		assert.Equal(t, int64(0), exists)
	}
}

package integration_test

import (
	"testing"

	"github.com/jbrinkman/valkey-ai-tasks/go/tests/utils"
	"github.com/stretchr/testify/assert"
)

// TestExample is a simple example integration test
func TestExample(t *testing.T) {
	// Use the test utilities
	ctx, req, cancel := utils.SetupTest(t)
	defer utils.CleanupTest(t, cancel)

	// This is just a placeholder test to verify the integration test structure
	assert.NotNil(t, ctx)
	req.True(true, "This test should always pass")
}

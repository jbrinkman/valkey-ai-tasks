package utils_test

import (
	"testing"
	"time"

	"github.com/jbrinkman/valkey-ai-tasks/go/tests/utils"
	"github.com/stretchr/testify/assert"
)

func TestTestContext(t *testing.T) {
	ctx, cancel := utils.TestContext(t)
	defer cancel()

	// Verify that the context has a deadline
	deadline, hasDeadline := ctx.Deadline()
	assert.True(t, hasDeadline)
	assert.WithinDuration(t, time.Now().Add(5*time.Second), deadline, 100*time.Millisecond)
}

func TestRequireTestify(t *testing.T) {
	req := utils.RequireTestify(t)
	
	// Simple test to verify the require instance works
	req.True(true, "This should pass")
}

func TestSetupTest(t *testing.T) {
	ctx, req, cancel := utils.SetupTest(t)
	defer utils.CleanupTest(t, cancel)

	// Verify that the context has a deadline
	_, hasDeadline := ctx.Deadline()
	assert.True(t, hasDeadline)
	
	// Verify that the require instance works
	req.NotNil(ctx, "Context should not be nil")
}

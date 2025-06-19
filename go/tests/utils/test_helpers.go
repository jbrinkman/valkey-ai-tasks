// Package utils provides testing utilities for the valkey-ai-tasks project
package utils

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestContext returns a context with timeout for use in tests
func TestContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// RequireTestify returns a testify require instance for cleaner assertions
// This is just a convenience function to make test code more readable
func RequireTestify(t *testing.T) *require.Assertions {
	t.Helper()
	return require.New(t)
}

// SetupTest is a helper function to set up common test requirements
// It returns a context with timeout and a testify require instance
func SetupTest(t *testing.T) (context.Context, *require.Assertions, context.CancelFunc) {
	t.Helper()
	ctx, cancel := TestContext(t)
	req := RequireTestify(t)
	return ctx, req, cancel
}

// CleanupTest is a helper function to clean up after tests
func CleanupTest(t *testing.T, cancel context.CancelFunc) {
	t.Helper()
	cancel()
}

// ParseHostPort splits a host:port string into separate host and port components
// Returns a slice with [host, port] or empty strings if parsing fails
func ParseHostPort(hostPort string) []string {
	parts := strings.Split(hostPort, ":")
	if len(parts) != 2 {
		return []string{"localhost", "6379"} // Default values
	}
	return parts
}

// ParseInt converts a string to an integer with a default value if parsing fails
func ParseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 6379 // Default port
	}
	return val
}

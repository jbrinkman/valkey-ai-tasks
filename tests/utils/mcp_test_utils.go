// Package utils provides testing utilities for the valkey-ai-tasks project
package utils

import (
	"net"
	"testing"

	"github.com/jbrinkman/valkey-ai-tasks/internal/mcp"
	"github.com/jbrinkman/valkey-ai-tasks/internal/storage"
	"github.com/stretchr/testify/require"
)

// GetRandomPort returns a random available port for testing
func GetRandomPort(t *testing.T) int {
	t.Helper()

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	require.NoError(t, err, "Failed to resolve TCP address")

	l, err := net.ListenTCP("tcp", addr)
	require.NoError(t, err, "Failed to listen on TCP address")
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port
}

// CreateTestMCPServer creates an MCP server for testing with a random port
func CreateTestMCPServer(t *testing.T, planRepo storage.PlanRepositoryInterface, taskRepo storage.TaskRepositoryInterface) (*mcp.MCPGoServer, int) {
	t.Helper()

	// Get a random port for the test server
	testPort := GetRandomPort(t)

	// Create MCP server with the repositories
	mcpServer := mcp.NewMCPGoServer(planRepo, taskRepo)

	return mcpServer, testPort
}

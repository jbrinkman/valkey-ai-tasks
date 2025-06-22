package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jbrinkman/valkey-ai-tasks/go/internal/storage"
	"github.com/jbrinkman/valkey-ai-tasks/go/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MCPServerTestSuite is a test suite for the MCP server
type MCPServerTestSuite struct {
	utils.RepositoryTestSuite
}

// TestMCPServerRandomPort tests that the MCP server uses a random port
func (s *MCPServerTestSuite) TestMCPServerRandomPort() {
	// Get repositories
	planRepo := s.GetPlanRepository()
	taskRepo := s.GetTaskRepository()

	// Create MCP server with random port
	mcpServer, port := utils.CreateTestMCPServer(s.T(), planRepo, taskRepo)
	
	// Verify the port is not the default 8080
	assert.NotEqual(s.T(), 8080, port, "Test MCP server should not use default port 8080")
	
	// Start the server in a goroutine
	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- mcpServer.Start(port)
	}()
	
	// Give the server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Verify the server is running on the specified port
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/mcp", port))
	if err == nil && resp != nil {
		defer resp.Body.Close()
		assert.Equal(s.T(), http.StatusOK, resp.StatusCode, "MCP server should respond on the random port")
	}
	
	// Check if the server started successfully
	select {
	case err := <-serverErrCh:
		require.NoError(s.T(), err, "MCP server should start without errors")
	default:
		// Server is still running, which is expected
	}
	
	// We don't need to stop the server explicitly as it will be stopped
	// when the test function exits and the goroutine is terminated
}

// TestMCPServerSuite runs the MCP server test suite
func TestMCPServerSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	suite.Run(t, new(MCPServerTestSuite))
}

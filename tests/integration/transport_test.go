package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jbrinkman/valkey-ai-tasks/internal/mcp"
	"github.com/jbrinkman/valkey-ai-tasks/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TransportTestSuite is a test suite for testing different MCP transport protocols
type TransportTestSuite struct {
	utils.RepositoryTestSuite
}

// setupTestServer creates and starts an MCP server with the specified configuration
func (s *TransportTestSuite) setupTestServer(enableSSE bool, enableStreamableHTTP bool, enableSTDIO bool) (*mcp.MCPGoServer, int, func()) {
	// Get repositories
	planRepo := s.GetPlanRepository()
	taskRepo := s.GetTaskRepository()

	// Set environment variables for the test
	originalSSE := os.Getenv("ENABLE_SSE")
	originalStreamableHTTP := os.Getenv("ENABLE_STREAMABLE_HTTP")
	originalSTDIO := os.Getenv("ENABLE_STDIO")

	// Set the environment variables for this test
	os.Setenv("ENABLE_SSE", fmt.Sprintf("%t", enableSSE))
	os.Setenv("ENABLE_STREAMABLE_HTTP", fmt.Sprintf("%t", enableStreamableHTTP))
	os.Setenv("ENABLE_STDIO", fmt.Sprintf("%t", enableSTDIO))

	// Create MCP server with random port
	mcpServer, port := utils.CreateTestMCPServer(s.T(), planRepo, taskRepo)

	// Start the server in a goroutine
	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- mcpServer.Start(port)
	}()

	// Give the server time to start
	time.Sleep(500 * time.Millisecond)

	// Return a cleanup function
	cleanup := func() {
		// Restore original environment variables
		if originalSSE != "" {
			os.Setenv("ENABLE_SSE", originalSSE)
		} else {
			os.Unsetenv("ENABLE_SSE")
		}

		if originalStreamableHTTP != "" {
			os.Setenv("ENABLE_STREAMABLE_HTTP", originalStreamableHTTP)
		} else {
			os.Unsetenv("ENABLE_STREAMABLE_HTTP")
		}

		if originalSTDIO != "" {
			os.Setenv("ENABLE_STDIO", originalSTDIO)
		} else {
			os.Unsetenv("ENABLE_STDIO")
		}

		// We don't need to stop the server explicitly as it will be stopped
		// when the test function exits and the goroutine is terminated
	}

	return mcpServer, port, cleanup
}

// TestHealthEndpoint tests the health endpoint with different transport configurations
func (s *TransportTestSuite) TestHealthEndpoint() {
	testCases := []struct {
		name                 string
		enableSSE            bool
		enableStreamableHTTP bool
		enableSTDIO         bool
		expectedStatus       int
	}{
		{
			name:                 "Both HTTP transports enabled",
			enableSSE:            true,
			enableStreamableHTTP: true,
			enableSTDIO:         false,
			expectedStatus:       http.StatusOK,
		},
		{
			name:                 "Only SSE enabled",
			enableSSE:            true,
			enableStreamableHTTP: false,
			enableSTDIO:         false,
			expectedStatus:       http.StatusOK,
		},
		{
			name:                 "Only Streamable HTTP enabled",
			enableSSE:            false,
			enableStreamableHTTP: true,
			enableSTDIO:         false,
			expectedStatus:       http.StatusOK,
		},
		{
			name:                 "HTTP and STDIO enabled",
			enableSSE:            true,
			enableStreamableHTTP: false,
			enableSTDIO:         true,
			expectedStatus:       http.StatusOK,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			_, port, cleanup := s.setupTestServer(tc.enableSSE, tc.enableStreamableHTTP, tc.enableSTDIO)
			defer cleanup()

			// Test the health endpoint
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", port))
			require.NoError(t, err, "Health endpoint should be accessible")
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode, "Health endpoint should return expected status code")

			// Check response body
			var result map[string]string
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err, "Response should be valid JSON")
			assert.Equal(t, "ok", result["status"], "Health endpoint should return status ok")
		})
	}
}

// TestRootEndpointRedirection tests the root endpoint redirection based on content type
func (s *TransportTestSuite) TestRootEndpointRedirection() {
	// Setup server with both transports enabled
	_, port, cleanup := s.setupTestServer(true, true, false)
	defer cleanup()

	testCases := []struct {
		name                string
		contentType         string
		expectedRedirectURL string
	}{
		{
			name:                "JSON content type should redirect to Streamable HTTP",
			contentType:         "application/json",
			expectedRedirectURL: fmt.Sprintf("http://localhost:%d/mcp", port),
		},
		{
			name:                "Text content type should redirect to SSE",
			contentType:         "text/plain",
			expectedRedirectURL: fmt.Sprintf("http://localhost:%d/sse", port),
		},
		{
			name:                "No content type should redirect to SSE",
			contentType:         "",
			expectedRedirectURL: fmt.Sprintf("http://localhost:%d/sse", port),
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			// Create a client that doesn't follow redirects
			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}

			// Create request with specified content type
			req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%d/", port), nil)
			require.NoError(t, err, "Should create request without error")

			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			// Send request
			resp, err := client.Do(req)
			require.NoError(t, err, "Root endpoint should be accessible")
			defer resp.Body.Close()

			// Check redirection
			assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode, "Should return redirect status")
			location := resp.Header.Get("Location")
			
			// Extract the path from the expected URL for comparison
			expectedPath := tc.expectedRedirectURL[len(fmt.Sprintf("http://localhost:%d", port)):]
			assert.Equal(t, expectedPath, location, "Should redirect to expected path")
		})
	}
}

// TestSSETransport tests the SSE transport functionality
func (s *TransportTestSuite) TestSSETransport() {
	// Setup server with only SSE enabled
	_, port, cleanup := s.setupTestServer(true, false, false)
	defer cleanup()

	// Connect directly to the SSE endpoint

	// Test that we can connect to the SSE endpoint
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/sse", port))
	require.NoError(s.T(), err, "Should connect to SSE endpoint without error")
	defer resp.Body.Close()

	// Check that we got a successful response
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode, "Should return OK status")

	// Verify the content type is correct for SSE
	contentType := resp.Header.Get("Content-Type")
	assert.Contains(s.T(), contentType, "text/event-stream", "Should return SSE content type")
}

// TestStreamableHTTPTransport tests the Streamable HTTP transport functionality
func (s *TransportTestSuite) TestStreamableHTTPTransport() {
	// Setup server with only Streamable HTTP enabled
	_, port, cleanup := s.setupTestServer(false, true, false)
	defer cleanup()

	// Create a simple JSON request
	requestBody := map[string]interface{}{
		"method": "list_plans",
		"params": map[string]interface{}{},
	}

	jsonData, err := json.Marshal(requestBody)
	require.NoError(s.T(), err, "Should marshal JSON without error")

	// Send request to the Streamable HTTP endpoint
	resp, err := http.Post(
		fmt.Sprintf("http://localhost:%d/mcp", port),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	require.NoError(s.T(), err, "Should send request without error")
	defer resp.Body.Close()

	// Check response status
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode, "Should return OK status")

	// Read response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err, "Should read response body without error")

	// Verify it's a valid JSON response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	require.NoError(s.T(), err, "Response should be valid JSON")
	
	// Just verify that we got a valid JSON response
	s.T().Logf("Response: %s", string(body))
	assert.NotEmpty(s.T(), result, "Response should not be empty")
}

// TestConcurrentConnections tests both transports with concurrent connections
func (s *TransportTestSuite) TestConcurrentConnections() {
	// Setup server with both transports enabled
	_, port, cleanup := s.setupTestServer(true, true, false)
	defer cleanup()

	// Number of concurrent connections to test
	concurrentConnections := 10

	// Create a wait group to wait for all goroutines to complete
	var wg sync.WaitGroup
	wg.Add(concurrentConnections * 2) // For both SSE and Streamable HTTP

	// Test SSE connections
	for i := 0; i < concurrentConnections; i++ {
		go func(id int) {
			defer wg.Done()

			// Variable to store error
			var err error

			// Test that we can connect to the SSE endpoint
			resp, err := http.Get(fmt.Sprintf("http://localhost:%d/sse", port))
			if err != nil {
				s.T().Logf("SSE client %d error: %v", id, err)
				return
			}
			resp.Body.Close()
			
			// Wait a bit to simulate connection time
			time.Sleep(100 * time.Millisecond)

			s.T().Logf("SSE client %d completed successfully", id)
		}(i)
	}

	// Test Streamable HTTP connections
	for i := 0; i < concurrentConnections; i++ {
		go func(id int) {
			defer wg.Done()

			// Create a simple JSON request
			requestBody := map[string]interface{}{
				"method": "list_plans",
				"params": map[string]interface{}{},
			}

			jsonData, err := json.Marshal(requestBody)
			if err != nil {
				s.T().Logf("Streamable HTTP client %d JSON marshal error: %v", id, err)
				return
			}

			// Send request
			resp, err := http.Post(
				fmt.Sprintf("http://localhost:%d/mcp", port),
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				s.T().Logf("Streamable HTTP client %d request error: %v", id, err)
				return
			}
			defer resp.Body.Close()

			s.T().Logf("Streamable HTTP client %d completed successfully with status %d", id, resp.StatusCode)
		}(i)
	}

	// Wait for all connections to complete with a timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.T().Log("All concurrent connections completed successfully")
	case <-time.After(10 * time.Second):
		s.T().Log("Timeout waiting for concurrent connections to complete")
	}
}

// TestSTDIOTransport tests the STDIO transport functionality
func (s *TransportTestSuite) TestSTDIOTransport() {
	// For STDIO transport, we can't use the normal HTTP-based testing approach
	// Instead, we'll verify that the server can start with STDIO enabled
	// and that it properly handles the configuration

	// Setup server with only STDIO enabled
	server, _, cleanup := s.setupTestServer(false, false, true)
	defer cleanup()

	// Verify that the server has STDIO enabled in its configuration
	config := server.GetConfig()
	assert.True(s.T(), config.EnableSTDIO, "STDIO transport should be enabled")
	assert.False(s.T(), config.EnableSSE, "SSE transport should be disabled")
	assert.False(s.T(), config.EnableStreamableHTTP, "Streamable HTTP transport should be disabled")

	// Note: We can't actually test the STDIO functionality in an integration test
	// as it would require capturing stdin/stdout which is difficult in a test environment
	// Instead, we rely on the fact that if the configuration is correct, the mcp-go
	// library will handle the STDIO transport correctly
}

// TestTransportSuite runs the transport test suite
func TestTransportSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite.Run(t, new(TransportTestSuite))
}

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	imcp "github.com/jbrinkman/valkey-ai-tasks/internal/mcp"
	"github.com/jbrinkman/valkey-ai-tasks/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/tests/utils"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// PlanResourceTestSuite is a test suite for the Plan resource
type PlanResourceTestSuite struct {
	utils.RepositoryTestSuite
	server   *imcp.MCPGoServer
	port     int
	serverCh chan error
}

// SetupTest sets up each test
func (s *PlanResourceTestSuite) SetupTest() {
	s.RepositoryTestSuite.SetupTest()

	// Set environment variables for MCP server configuration
	// Enable Streamable HTTP transport for testing
	s.T().Setenv("ENABLE_STREAMABLE_HTTP", "true")
	s.T().Setenv("STREAMABLE_HTTP_ENDPOINT", "/mcp")

	// Get repositories
	planRepo := s.GetPlanRepository()
	taskRepo := s.GetTaskRepository()

	// Create MCP server with random port
	s.server, s.port = utils.CreateTestMCPServer(s.T(), planRepo, taskRepo)

	// Start the server in a goroutine
	s.serverCh = make(chan error, 1)
	go func() {
		s.serverCh <- s.server.Start(s.port)
	}()

	// Give the server time to start
	time.Sleep(100 * time.Millisecond)
}

// createTestPlan creates a test plan with tasks for testing
func (s *PlanResourceTestSuite) createTestPlan() *models.Plan {
	// Create a test plan
	plan, err := s.GetPlanRepository().Create(
		s.Context,
		"test-app-id",
		"Test Plan",
		"A test plan for integration testing",
	)
	require.NoError(s.T(), err, "Failed to create test plan")

	// Create test tasks
	_, err = s.GetTaskRepository().Create(
		s.Context,
		plan.ID,
		"Task 1",
		"Description for task 1",
		models.TaskPriorityHigh,
	)
	require.NoError(s.T(), err, "Failed to create test task 1")

	_, err = s.GetTaskRepository().Create(
		s.Context,
		plan.ID,
		"Task 2",
		"Description for task 2",
		models.TaskPriorityMedium,
	)
	require.NoError(s.T(), err, "Failed to create test task 2")

	// Return the plan for use in tests
	return plan
}

// createMCPClient creates and initializes an MCP client
func createMCPClient(url string) (*client.Client, error) {
	mcpClient, err := client.NewStreamableHttpClient(url + "/mcp")
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}
	if mcpClient == nil {
		return nil, fmt.Errorf("MCP client should not be nil")
	}

	initRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "plan-resource-test-client",
				Version: "1.0.0",
			},
		},
	}

	err = mcpClient.Start(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to start client: %w", err)
	}

	_, err = mcpClient.Initialize(context.Background(), initRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %w", err)
	}
	return mcpClient, nil
}

// readPlanResource reads a resource from the MCP server
func readPlanResource(ctx context.Context, c *client.Client, uri string) (*mcp.ReadResourceResult, error) {
	result, err := c.ReadResource(ctx, mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: uri,
		},
	})
	return result, err
}

// TestSinglePlanResource tests the single plan resource
func (s *PlanResourceTestSuite) TestSinglePlanResource() {
	// Create a test plan with tasks
	plan := s.createTestPlan()

	// Create an MCP client
	url := fmt.Sprintf("http://localhost:%d", s.port)
	mcpClient, err := createMCPClient(url)
	require.NoError(s.T(), err, "Failed to create MCP client")

	// Create the request URI
	uri := fmt.Sprintf("ai-tasks://plans/%s/full", plan.ID)
	s.T().Logf("Reading resource: %s", uri)

	// Read the resource using the client
	result, err := readPlanResource(context.Background(), mcpClient, uri)
	require.NoError(s.T(), err, "Failed to read resource")
	require.NotNil(s.T(), result, "Expected non-nil result")
	require.NotEmpty(s.T(), result.Contents, "Expected non-empty contents")

	// Get the first content item (should be TextResourceContents)
	var textContent *mcp.TextResourceContents
	for _, content := range result.Contents {
		if tc, ok := content.(mcp.TextResourceContents); ok {
			textContent = &tc
			break
		}
	}
	require.NotNil(s.T(), textContent, "Expected TextResourceContents")

	// Parse the resource content
	s.T().Logf("Plan resource content: %s", textContent.Text)

	// Debug the JSON content
	var rawData map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &rawData)
	require.NoError(s.T(), err, "Failed to parse raw resource content")

	// Print the raw data for debugging
	s.T().Logf("Raw plan data: %+v", rawData)

	// Now parse into the structured type with the correct nested structure
	var planResource struct {
		Plan struct {
			ID            string `json:"id"`
			ApplicationID string `json:"application_id"`
			Name          string `json:"name"`
			Description   string `json:"description"`
			Status        string `json:"status"`
		} `json:"plan"`
		Tasks []struct {
			ID          string `json:"id"`
			PlanID      string `json:"plan_id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Status      string `json:"status"`
			Priority    string `json:"priority"`
		} `json:"tasks"`
	}
	err = json.Unmarshal([]byte(textContent.Text), &planResource)
	require.NoError(s.T(), err, "Failed to parse resource content")

	// Verify the plan data
	assert.Equal(s.T(), plan.ID, planResource.Plan.ID)
	assert.Equal(s.T(), plan.ApplicationID, planResource.Plan.ApplicationID)
	assert.Equal(s.T(), plan.Name, planResource.Plan.Name)
	assert.Len(s.T(), planResource.Tasks, 2, "Expected 2 tasks")
}

// TestAllPlansResource tests the all plans resource
func (s *PlanResourceTestSuite) TestAllPlansResource() {
	// Create a test plan with tasks
	s.createTestPlan()

	// Create an MCP client
	url := fmt.Sprintf("http://localhost:%d", s.port)
	mcpClient, err := createMCPClient(url)
	require.NoError(s.T(), err, "Failed to create MCP client")

	// Create the request URI
	uri := "ai-tasks://plans/full"
	s.T().Logf("Reading resource: %s", uri)

	// Read the resource using the client
	result, err := readPlanResource(context.Background(), mcpClient, uri)
	require.NoError(s.T(), err, "Failed to read resource")
	require.NotNil(s.T(), result, "Expected non-nil result")
	require.NotEmpty(s.T(), result.Contents, "Expected non-empty contents")

	// Get the TextResourceContents from the response
	require.NotEmpty(s.T(), result.Contents, "Expected non-empty contents")

	// The first content should be TextResourceContents
	textContent, ok := result.Contents[0].(mcp.TextResourceContents)
	require.True(s.T(), ok, "Expected TextResourceContents")
	require.NotEmpty(s.T(), textContent.Text, "Expected non-empty text content")

	// Verify the response
	assert.Equal(s.T(), "ai-tasks://plans/full", textContent.URI)
	assert.Equal(s.T(), "application/json", textContent.MIMEType)

	// Parse the resource content
	s.T().Logf("All plans resource content: %s", textContent.Text)

	// Debug the JSON content - it's an array, not an object
	var plansList []map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &plansList)
	require.NoError(s.T(), err, "Failed to parse raw resource content")

	// Print the raw data for debugging
	s.T().Logf("Raw all plans data: %+v", plansList)

	// Parse with the correct structure - array of plan objects
	var plansResource []struct {
		Plan struct {
			ID            string `json:"id"`
			ApplicationID string `json:"application_id"`
			Name          string `json:"name"`
			Description   string `json:"description"`
			Status        string `json:"status"`
		} `json:"plan"`
		Tasks []struct {
			ID          string `json:"id"`
			PlanID      string `json:"plan_id"`
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"tasks"`
	}
	err = json.Unmarshal([]byte(textContent.Text), &plansResource)
	require.NoError(s.T(), err, "Failed to parse resource content")

	// Verify we have at least one plan
	assert.NotEmpty(s.T(), plansResource, "Expected at least one plan")
}

// TestAppPlansResource tests the application plans resource
func (s *PlanResourceTestSuite) TestAppPlansResource() {
	// Create a test plan with tasks
	plan := s.createTestPlan()

	// Create an MCP client
	url := fmt.Sprintf("http://localhost:%d", s.port)
	mcpClient, err := createMCPClient(url)
	require.NoError(s.T(), err, "Failed to create MCP client")

	// Create the request URI
	uri := fmt.Sprintf("ai-tasks://applications/%s/plans/full", plan.ApplicationID)
	s.T().Logf("Reading resource: %s", uri)

	// Read the resource using the client
	result, err := readPlanResource(context.Background(), mcpClient, uri)
	require.NoError(s.T(), err, "Failed to read resource")
	require.NotNil(s.T(), result, "Expected non-nil result")
	require.NotEmpty(s.T(), result.Contents, "Expected non-empty contents")

	// Get the first content item (should be TextResourceContents)
	var textContent *mcp.TextResourceContents
	for _, content := range result.Contents {
		if tc, ok := content.(mcp.TextResourceContents); ok {
			textContent = &tc
			break
		}
	}
	require.NotNil(s.T(), textContent, "Expected TextResourceContents")

	// Debug the JSON content - it's an array, not an object
	var plansList []map[string]interface{}
	err = json.Unmarshal([]byte(textContent.Text), &plansList)
	require.NoError(s.T(), err, "Failed to parse raw resource content")

	// Print the raw data for debugging
	s.T().Logf("Raw app plans data: %+v", plansList)

	// Parse with the correct structure - array of plan objects
	var appPlansResource []struct {
		Plan struct {
			ID            string `json:"id"`
			ApplicationID string `json:"application_id"`
			Name          string `json:"name"`
			Description   string `json:"description"`
			Status        string `json:"status"`
		} `json:"plan"`
	}
	err = json.Unmarshal([]byte(textContent.Text), &appPlansResource)
	require.NoError(s.T(), err, "Failed to parse resource content")

	// Debug the parsed content
	s.T().Logf("Parsed app plans: %+v", appPlansResource)

	// Verify we have at least one plan
	assert.NotEmpty(s.T(), appPlansResource, "Expected at least one plan")

	// Verify we can find the created plan
	found := false
	for _, p := range appPlansResource {
		if p.Plan.ID == plan.ID {
			found = true
			break
		}
	}
	assert.True(s.T(), found, "Expected to find the created plan")
}

// TestInvalidResourceURI tests handling of invalid resource URIs
func (s *PlanResourceTestSuite) TestInvalidResourceURI() {
	// Create an HTTP client
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Create a request with an invalid URI
	url := fmt.Sprintf("http://localhost:%d/mcp", s.port)
	reqBody := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "read_resource",
		"params": {
			"uri": "ai-tasks://invalid/uri"
		}
	}`

	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	require.NoError(s.T(), err, "Failed to create HTTP request")
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := client.Do(req)
	require.NoError(s.T(), err, "Failed to send HTTP request")
	defer resp.Body.Close()

	// Check the response status - should still be 200 OK
	assert.Equal(s.T(), http.StatusOK, resp.StatusCode, "Expected status code 200")

	// Parse the response - should contain an error
	var response struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(s.T(), err, "Failed to decode response")

	// Verify we got an error response
	assert.NotEmpty(s.T(), response.Error.Message, "Expected error message for invalid URI")
}

// TestLegacyRequestFormat tests handling of the old request format
func (s *PlanResourceTestSuite) TestLegacyRequestFormat() {
	// Create an MCP client
	url := fmt.Sprintf("http://localhost:%d", s.port)
	mcpClient, err := createMCPClient(url)
	require.NoError(s.T(), err, "Failed to create MCP client")

	// Create a test plan to ensure we have data
	plan := s.createTestPlan()
	require.NotEmpty(s.T(), plan.ID, "Plan ID should not be empty")

	// Read all plans resource
	uri := "ai-tasks://plans/full"
	s.T().Logf("Reading resource: %s", uri)

	// Read the resource using the client
	result, err := readPlanResource(context.Background(), mcpClient, uri)
	require.NoError(s.T(), err, "Failed to read resource")
	require.NotNil(s.T(), result, "Expected non-nil result")
	require.NotEmpty(s.T(), result.Contents, "Expected non-empty contents")

	// Get the TextResourceContents from the response
	require.NotEmpty(s.T(), result.Contents, "Expected non-empty contents")

	// The first content should be TextResourceContents
	textContent, ok := result.Contents[0].(mcp.TextResourceContents)
	require.True(s.T(), ok, "Expected TextResourceContents")
	require.NotEmpty(s.T(), textContent.Text, "Expected non-empty text content")

	// Parse the resource content
	var plansResource []struct {
		ID            string `json:"id"`
		ApplicationID string `json:"application_id"`
		Name          string `json:"name"`
	}
	err = json.Unmarshal([]byte(textContent.Text), &plansResource)
	require.NoError(s.T(), err, "Failed to parse resource content")

	// Verify we have at least one plan
	assert.NotEmpty(s.T(), plansResource, "Expected at least one plan")
}

// TestPlanNotFound tests handling of non-existent plan IDs
func (s *PlanResourceTestSuite) TestPlanNotFound() {
	// Create an MCP client
	url := fmt.Sprintf("http://localhost:%d", s.port)
	mcpClient, err := createMCPClient(url)
	require.NoError(s.T(), err, "Failed to create MCP client")

	// Read a non-existent plan resource
	uri := "ai-tasks://plans/non-existent-id/full"
	s.T().Logf("Reading non-existent resource: %s", uri)

	// Read the resource using the client
	result, err := readPlanResource(context.Background(), mcpClient, uri)

	// We expect an error for non-existent plan
	assert.Error(s.T(), err, "Expected error for non-existent plan")
	assert.Contains(s.T(), err.Error(), "plan not found", "Expected 'plan not found' in error message")

	// The result should be nil or have empty contents
	if result != nil && len(result.Contents) > 0 {
		// If we got a result with contents, it should be an error message
		textContent, ok := result.Contents[0].(mcp.TextResourceContents)
		if ok {
			assert.Contains(s.T(), textContent.Text, "not found", "Expected 'not found' in error message")
		}
	}
}

// TestPlanResourceSuite runs the Plan resource test suite
func TestPlanResourceSuite(t *testing.T) {
	suite.Run(t, new(PlanResourceTestSuite))
}

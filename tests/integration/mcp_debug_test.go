package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	imcp "github.com/jbrinkman/valkey-ai-tasks/internal/mcp"
	"github.com/jbrinkman/valkey-ai-tasks/tests/utils"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MCPDebugTestSuite is a test suite for debugging MCP server issues
type MCPDebugTestSuite struct {
	utils.RepositoryTestSuite
	server   *imcp.MCPGoServer
	port     int
	serverCh chan error
}

// SetupTest sets up the individual tests
func (s *MCPDebugTestSuite) SetupTest() {
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

// TestMCPDebug tests the MCP server with debug output
func (s *MCPDebugTestSuite) TestMCPDebug() {
	// Create an MCP client
	url := fmt.Sprintf("http://localhost:%d", s.port)
	s.T().Logf("Creating MCP client for URL: %s", url)

	// Create StreamableHTTP client
	mcpClient, err := getStreamableClient_Orig(s, url)
	require.NoError(s.T(), err, "Failed to create MCP client")

	// Create the request
	s.T().Logf("Reading resource: ai-tasks://plans/full")

	// Read the resource using the client
	resultRead, err := readResource_Orig(context.Background(), mcpClient, "ai-tasks://plans/full")
	require.NoError(s.T(), err, "Failed to read resource")

	if resultRead != nil && resultRead.Contents != nil {
		s.T().Logf("Got %d content items", len(resultRead.Contents))

		for i, content := range resultRead.Contents {
			// Type assertion to access specific implementations
			switch c := content.(type) {
			case mcp.TextResourceContents:
				s.T().Logf("Content %d: URI=%s, MIME=%s, Type=Text, Length=%d",
					i, c.URI, c.MIMEType, len(c.Text))
			case mcp.BlobResourceContents:
				s.T().Logf("Content %d: URI=%s, MIME=%s, Type=Blob, Length=%d",
					i, c.URI, c.MIMEType, len(c.Blob))
			default:
				s.T().Logf("Content %d: Unknown type %T", i, content)
			}
		}
	}

	// Assert that we got a successful response
	require.NoError(s.T(), err, "Failed to read resource")
	require.NotNil(s.T(), resultRead, "Expected non-nil result")
	require.NotEmpty(s.T(), resultRead.Contents, "Expected non-empty contents")
}

func getStreamableClient_Orig(s *MCPDebugTestSuite, url string) (*client.Client, error) {
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
				Name:    "debug-test-client",
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

func readResource_Orig(ctx context.Context, c *client.Client, uri string) (*mcp.ReadResourceResult, error) {
	result, err := c.ReadResource(ctx, mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: uri,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read resource %s: %w", uri, err)
	}

	return result, nil
}

// TestMCPDebugSuite runs the debug test suite
func TestMCPDebugSuite(t *testing.T) {
	suite.Run(t, new(MCPDebugTestSuite))
}

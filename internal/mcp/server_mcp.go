package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/server"

	"github.com/jbrinkman/valkey-ai-tasks/internal/storage"
)

// ServerConfig holds configuration for the MCP server
type ServerConfig struct {
	// EnableSSE controls whether the SSE transport is enabled
	EnableSSE bool
	// SSEEndpoint is the endpoint path for SSE transport
	SSEEndpoint string
	// SSEKeepAlive controls whether keep-alive is enabled for SSE
	SSEKeepAlive bool
	// SSEKeepAliveInterval is the interval for SSE keep-alive messages in seconds
	SSEKeepAliveInterval int

	// EnableStreamableHTTP controls whether the Streamable HTTP transport is enabled
	EnableStreamableHTTP bool
	// StreamableHTTPEndpoint is the endpoint path for Streamable HTTP transport
	StreamableHTTPEndpoint string
	// StreamableHTTPHeartbeatInterval is the interval for Streamable HTTP heartbeat messages in seconds
	StreamableHTTPHeartbeatInterval int
	// StreamableHTTPStateless controls whether the Streamable HTTP transport is stateless
	StreamableHTTPStateless bool

	// EnableSTDIO controls whether the STDIO transport is enabled
	EnableSTDIO bool
	// STDIOErrorLog controls whether to log errors to stderr
	STDIOErrorLog bool

	// ServerReadTimeout is the maximum duration for reading the entire request in seconds
	ServerReadTimeout int
	// ServerWriteTimeout is the maximum duration for writing the response in seconds
	ServerWriteTimeout int
}

// MCPGoServer wraps the mark3labs/mcp-go server implementation
type MCPGoServer struct {
	server   *server.MCPServer
	config   ServerConfig
	planRepo storage.PlanRepositoryInterface
	taskRepo storage.TaskRepositoryInterface
}

// NewMCPGoServer creates a new MCP server using the mark3labs/mcp-go library
func NewMCPGoServer(planRepo storage.PlanRepositoryInterface, taskRepo storage.TaskRepositoryInterface) *MCPGoServer {
	// Create a new MCP server
	s := server.NewMCPServer(
		"Valkey Feature Planning & Task Management",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	// Get configuration from environment variables
	config := getServerConfigFromEnv()

	mcpServer := &MCPGoServer{
		server:   s,
		config:   config,
		planRepo: planRepo,
		taskRepo: taskRepo,
	}

	// Register all tools
	mcpServer.registerTools()

	// Register all resources
	mcpServer.registerResources()

	return mcpServer
}

// getServerConfigFromEnv reads server configuration from environment variables
func getServerConfigFromEnv() ServerConfig {
	// Default configuration
	config := ServerConfig{
		// SSE configuration
		EnableSSE:            true,
		SSEEndpoint:          "/sse",
		SSEKeepAlive:         true,
		SSEKeepAliveInterval: 30,

		// Streamable HTTP configuration
		EnableStreamableHTTP:            false,
		StreamableHTTPEndpoint:          "/mcp",
		StreamableHTTPHeartbeatInterval: 30,
		StreamableHTTPStateless:         false,

		// STDIO configuration
		EnableSTDIO:   false,
		STDIOErrorLog: true,

		// Server configuration
		ServerReadTimeout:  60,
		ServerWriteTimeout: 60,
	}

	// SSE configuration from environment variables
	if val := os.Getenv("ENABLE_SSE"); val != "" {
		config.EnableSSE = strings.ToLower(val) == "true"
	}

	if val := os.Getenv("SSE_ENDPOINT"); val != "" {
		config.SSEEndpoint = val
	}

	if val := os.Getenv("SSE_KEEP_ALIVE"); val != "" {
		config.SSEKeepAlive = strings.ToLower(val) == "true"
	}

	if val := os.Getenv("SSE_KEEP_ALIVE_INTERVAL"); val != "" {
		if interval, err := strconv.Atoi(val); err == nil && interval > 0 {
			config.SSEKeepAliveInterval = interval
		}
	}

	// Streamable HTTP configuration from environment variables
	if val := os.Getenv("ENABLE_STREAMABLE_HTTP"); val != "" {
		config.EnableStreamableHTTP = strings.ToLower(val) == "true"
	}

	if val := os.Getenv("STREAMABLE_HTTP_ENDPOINT"); val != "" {
		config.StreamableHTTPEndpoint = val
	}

	if val := os.Getenv("STREAMABLE_HTTP_HEARTBEAT_INTERVAL"); val != "" {
		if interval, err := strconv.Atoi(val); err == nil && interval > 0 {
			config.StreamableHTTPHeartbeatInterval = interval
		}
	}

	if val := os.Getenv("STREAMABLE_HTTP_STATELESS"); val != "" {
		config.StreamableHTTPStateless = strings.ToLower(val) == "true"
	}

	// STDIO configuration from environment variables
	if val := os.Getenv("ENABLE_STDIO"); val != "" {
		config.EnableSTDIO = strings.ToLower(val) == "true"
	}

	if val := os.Getenv("STDIO_ERROR_LOG"); val != "" {
		config.STDIOErrorLog = strings.ToLower(val) == "true"
	}

	// Server configuration from environment variables
	if val := os.Getenv("SERVER_READ_TIMEOUT"); val != "" {
		if timeout, err := strconv.Atoi(val); err == nil && timeout > 0 {
			config.ServerReadTimeout = timeout
		}
	}

	if val := os.Getenv("SERVER_WRITE_TIMEOUT"); val != "" {
		if timeout, err := strconv.Atoi(val); err == nil && timeout > 0 {
			config.ServerWriteTimeout = timeout
		}
	}

	log.Printf("Server configuration: %+v", config)

	return config
}

// transportSelectionHandler handles requests to the root path and selects the appropriate transport
// based on the request's content-type header
func (s *MCPGoServer) transportSelectionHandler(w http.ResponseWriter, r *http.Request) {
	// If the request path is not root, return 404
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Check content-type header for transport selection
	contentType := r.Header.Get("Content-Type")

	// If STDIO is enabled, add it to the response information
	stdioEnabled := s.config.EnableSTDIO

	// Default to SSE if multiple transports are enabled
	if s.config.EnableSSE && s.config.EnableStreamableHTTP {
		// If content-type indicates JSON, use Streamable HTTP
		if strings.Contains(contentType, "application/json") {
			http.Redirect(w, r, s.config.StreamableHTTPEndpoint, http.StatusTemporaryRedirect)
			return
		}
		// Otherwise default to SSE
		http.Redirect(w, r, s.config.SSEEndpoint, http.StatusTemporaryRedirect)
		return
	}

	// If only one HTTP transport is enabled, redirect to it
	if s.config.EnableSSE {
		http.Redirect(w, r, s.config.SSEEndpoint, http.StatusTemporaryRedirect)
		return
	}

	if s.config.EnableStreamableHTTP {
		http.Redirect(w, r, s.config.StreamableHTTPEndpoint, http.StatusTemporaryRedirect)
		return
	}

	// If only STDIO is enabled, show information about it
	if stdioEnabled {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "ok",
			"message":    "This server is configured for STDIO transport only. HTTP endpoints are not available.",
			"transports": []string{"stdio"},
		})
		return
	}

	// If we get here, no transports are enabled (should not happen due to earlier check)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	json.NewEncoder(w).Encode(map[string]string{"error": "No transport protocols are enabled on this server"})
}

// GetConfig returns the current server configuration
func (s *MCPGoServer) GetConfig() ServerConfig {
	return s.config
}

// Start starts the MCP server using the configured transports
func (s *MCPGoServer) Start(port int) error {
	log.Printf("Starting MCP server on port %d", port)

	// Check if at least one transport is enabled
	if !s.config.EnableSSE && !s.config.EnableStreamableHTTP && !s.config.EnableSTDIO {
		return fmt.Errorf("no transport protocols enabled, enable at least one of SSE, Streamable HTTP, or STDIO")
	}

	// If STDIO is enabled, handle it separately as it's not compatible with HTTP server
	if s.config.EnableSTDIO {
		log.Printf("Enabling STDIO transport")

		// Only run STDIO if it's the only transport enabled
		if !s.config.EnableSSE && !s.config.EnableStreamableHTTP {
			// Configure STDIO options
			var stdioOptions []server.StdioOption

			// Add error logger if enabled
			if s.config.STDIOErrorLog {
				stdioOptions = append(stdioOptions, server.WithErrorLogger(log.Default()))
			}

			// Start STDIO server - this will block until terminated
			return server.ServeStdio(s.server, stdioOptions...)
		}
	}

	// Create a new HTTP server mux for routing
	mux := http.NewServeMux()

	// Configure SSE transport if enabled
	if s.config.EnableSSE {
		log.Printf("Enabling SSE transport at endpoint: %s", s.config.SSEEndpoint)

		// Create SSE server with configuration options
		sseOptions := []server.SSEOption{
			server.WithSSEEndpoint(s.config.SSEEndpoint),
			server.WithKeepAlive(s.config.SSEKeepAlive),
		}

		// Add keep-alive interval if keep-alive is enabled
		if s.config.SSEKeepAlive && s.config.SSEKeepAliveInterval > 0 {
			sseOptions = append(sseOptions,
				server.WithKeepAliveInterval(time.Duration(s.config.SSEKeepAliveInterval)*time.Second),
			)
		}

		sseServer := server.NewSSEServer(s.server, sseOptions...)
		mux.Handle(s.config.SSEEndpoint, sseServer)
	}

	// Configure Streamable HTTP transport if enabled
	if s.config.EnableStreamableHTTP {
		log.Printf("Enabling Streamable HTTP transport at endpoint: %s", s.config.StreamableHTTPEndpoint)

		// Create Streamable HTTP server with configuration options
		streamableOptions := []server.StreamableHTTPOption{
			server.WithEndpointPath(s.config.StreamableHTTPEndpoint),
			server.WithStateLess(s.config.StreamableHTTPStateless),
		}

		// Add heartbeat interval if configured
		if s.config.StreamableHTTPHeartbeatInterval > 0 {
			streamableOptions = append(streamableOptions,
				server.WithHeartbeatInterval(time.Duration(s.config.StreamableHTTPHeartbeatInterval)*time.Second),
			)
		}

		streamableServer := server.NewStreamableHTTPServer(s.server, streamableOptions...)
		mux.Handle(s.config.StreamableHTTPEndpoint, streamableServer)
	}

	// Add a simple health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Add a root handler for transport selection based on content-type
	mux.HandleFunc("/", s.transportSelectionHandler)

	// Create and start the HTTP server with timeouts
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  time.Duration(s.config.ServerReadTimeout) * time.Second,
		WriteTimeout: time.Duration(s.config.ServerWriteTimeout) * time.Second,
	}

	return httpServer.ListenAndServe()
}

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/jbrinkman/valkey-ai-tasks/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/internal/storage"
	"github.com/jbrinkman/valkey-ai-tasks/internal/utils/markdown"
)

// TransportType represents the type of transport protocol
type TransportType string

const (
	// TransportSSE represents Server-Sent Events transport
	TransportSSE TransportType = "sse"
	// TransportStreamableHTTP represents Streamable HTTP transport
	TransportStreamableHTTP TransportType = "streamable-http"
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

	return mcpServer
}

// getServerConfigFromEnv reads server configuration from environment variables
func getServerConfigFromEnv() ServerConfig {
	// Default configuration
	config := ServerConfig{
		// SSE configuration
		EnableSSE:             true,
		SSEEndpoint:           "/sse",
		SSEKeepAlive:          true,
		SSEKeepAliveInterval:  15,

		// Streamable HTTP configuration
		EnableStreamableHTTP:          false,
		StreamableHTTPEndpoint:        "/mcp",
		StreamableHTTPHeartbeatInterval: 30,
		StreamableHTTPStateless:       false,

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

	// Default to SSE if both transports are enabled
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

	// If only one transport is enabled, redirect to it
	if s.config.EnableSSE {
		http.Redirect(w, r, s.config.SSEEndpoint, http.StatusTemporaryRedirect)
		return
	}

	if s.config.EnableStreamableHTTP {
		http.Redirect(w, r, s.config.StreamableHTTPEndpoint, http.StatusTemporaryRedirect)
		return
	}

	// If we get here, no transports are enabled (should not happen due to earlier check)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	json.NewEncoder(w).Encode(map[string]string{"error": "No transport protocols are enabled on this server"})
}

// Start starts the MCP server using the configured transports
func (s *MCPGoServer) Start(port int) error {
	log.Printf("Starting MCP server on port %d", port)

	// Check if at least one transport is enabled
	if !s.config.EnableSSE && !s.config.EnableStreamableHTTP {
		return fmt.Errorf("no transport protocols enabled, enable at least one of SSE or Streamable HTTP")
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
				server.WithKeepAliveInterval(time.Duration(s.config.SSEKeepAliveInterval) * time.Second),
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
				server.WithHeartbeatInterval(time.Duration(s.config.StreamableHTTPHeartbeatInterval) * time.Second),
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

// registerTools registers all the task management tools with the MCP server
func (s *MCPGoServer) registerTools() {
	// Plan tools
	s.registerCreatePlanTool()
	s.registerGetPlanTool()
	s.registerListPlansTool()
	s.registerListPlansByApplicationTool()
	s.registerUpdatePlanTool()
	s.registerDeletePlanTool()
	s.registerUpdatePlanStatusTool()
	s.registerListPlansByStatusTool()

	// Plan notes tools
	s.registerUpdatePlanNotesTool()
	s.registerGetPlanNotesTool()

	// Task tools
	s.registerCreateTaskTool()
	s.registerGetTaskTool()
	s.registerListTasksByPlanTool()
	s.registerListTasksByStatusTool()
	s.registerListTasksByPlanAndStatusTool() // New tool
	s.registerUpdateTaskTool()
	s.registerDeleteTaskTool()
	s.registerBulkCreateTasksTool()
	s.registerReorderTaskTool()
	s.registerListOrphanedTasksTool()

	// Task notes tools
	s.registerUpdateTaskNotesTool()
	s.registerGetTaskNotesTool()
}

// Plan tools implementation

func (s *MCPGoServer) registerCreatePlanTool() {
	tool := mcp.NewTool("create_plan",
		mcp.WithDescription("Create a new plan for planning and organizing a feature or initiative"),
		mcp.WithString("application_id",
			mcp.Required(),
			mcp.Description("The application ID this plan belongs to"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the feature or initiative being planned"),
		),
		mcp.WithString("description",
			mcp.Description("Detailed description of the feature's goals, requirements, and scope (optional)"),
		),
		mcp.WithString("notes",
			mcp.Description("Initial Markdown-formatted notes for the plan (optional)"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		applicationID, err := request.RequireString("application_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		name, err := request.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		description := request.GetString("description", "no description provided")
		notes := request.GetString("notes", "")

		// Create the plan
		plan, err := s.planRepo.Create(ctx, applicationID, name, description)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create plan: %v", err)), nil
		}

		// If notes were provided, validate, format and update them
		if notes != "" {
			// Import markdown utilities
			if err := markdown.Validate(notes); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid notes format: %v", err)), nil
			}

			// Sanitize and format the notes
			notes = markdown.Sanitize(notes)
			notes = markdown.Format(notes)

			err = s.planRepo.UpdateNotes(ctx, plan.ID, notes)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to set initial notes: %v", err)), nil
			}

			// Refresh plan to include notes
			plan, err = s.planRepo.Get(ctx, plan.ID)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to refresh plan: %v", err)), nil
			}
		}

		planJson, err := json.Marshal(plan)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal plan: %v", err)), nil
		}
		return mcp.NewToolResultText(string(planJson)), nil
	})
}

func (s *MCPGoServer) registerGetPlanTool() {
	tool := mcp.NewTool("get_plan",
		mcp.WithDescription("Retrieve details about a specific feature planning plan"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Plan ID"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		plan, err := s.planRepo.Get(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get plan: %v", err)), nil
		}

		planJson, err := json.Marshal(plan)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal plan: %v", err)), nil
		}
		return mcp.NewToolResultText(string(planJson)), nil
	})
}

func (s *MCPGoServer) registerListPlansTool() {
	tool := mcp.NewTool("list_plans",
		mcp.WithDescription("List all available feature planning plans"),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		plans, err := s.planRepo.List(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list plans: %v", err)), nil
		}

		plansJson, err := json.Marshal(plans)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal plans: %v", err)), nil
		}
		return mcp.NewToolResultText(string(plansJson)), nil
	})
}

func (s *MCPGoServer) registerListPlansByApplicationTool() {
	tool := mcp.NewTool("list_plans_by_application",
		mcp.WithDescription("List all feature planning plans for a specific application"),
		mcp.WithString("application_id",
			mcp.Required(),
			mcp.Description("Application ID to filter plans by"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		applicationID, err := request.RequireString("application_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		plans, err := s.planRepo.ListByApplication(ctx, applicationID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list plans by application: %v", err)), nil
		}

		plansJson, err := json.Marshal(plans)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal plans: %v", err)), nil
		}
		return mcp.NewToolResultText(string(plansJson)), nil
	})
}

// validatePlanStatus checks if the provided status is a valid plan status
func validatePlanStatus(status models.PlanStatus) error {
	if status != models.PlanStatusNew &&
		status != models.PlanStatusInProgress &&
		status != models.PlanStatusCompleted &&
		status != models.PlanStatusCancelled {
		return fmt.Errorf("invalid status: %s", status)
	}
	return nil
}

func (s *MCPGoServer) registerUpdatePlanStatusTool() {
	tool := mcp.NewTool("update_plan_status",
		mcp.WithDescription("Update the status of a plan"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Plan ID"),
		),
		mcp.WithString("status",
			mcp.Required(),
			mcp.Description("New status value (new, inprogress, completed, cancelled)"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		statusStr, err := request.RequireString("status")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Validate status
		status := models.PlanStatus(statusStr)
		if err := validatePlanStatus(status); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get the existing plan
		plan, err := s.planRepo.Get(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get plan: %v", err)), nil
		}

		// Update status
		plan.Status = status
		plan.UpdatedAt = time.Now()

		// Save the updated plan
		err = s.planRepo.Update(ctx, plan)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update plan: %v", err)), nil
		}

		planJson, err := json.Marshal(plan)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal plan: %v", err)), nil
		}
		return mcp.NewToolResultText(string(planJson)), nil
	})
}

func (s *MCPGoServer) registerUpdatePlanTool() {
	tool := mcp.NewTool("update_plan",
		mcp.WithDescription("Update the details or scope of a feature planning plan"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Plan ID"),
		),
		mcp.WithString("name",
			mcp.Description("New plan name (optional)"),
		),
		mcp.WithString("description",
			mcp.Description("New plan description (optional)"),
		),
		mcp.WithString("notes",
			mcp.Description("New Markdown-formatted notes (optional)"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get the existing plan
		plan, err := s.planRepo.Get(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get plan: %v", err)), nil
		}

		// Update fields if provided
		name := request.GetString("name", plan.Name)
		if name != plan.Name {
			plan.Name = name
		}

		description := request.GetString("description", plan.Description)
		if description != plan.Description {
			plan.Description = description
		}

		// Check if notes are provided
		notes := request.GetString("notes", "")
		if notes != "" {
			// Validate and format the markdown content
			err = markdown.Validate(notes)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid notes format: %v", err)), nil
			}

			// Sanitize and format the notes
			notes = markdown.Sanitize(notes)
			notes = markdown.Format(notes)

			// Update notes separately using the dedicated method
			err = s.planRepo.UpdateNotes(ctx, id, notes)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to update notes: %v", err)), nil
			}
			// Update plan.Notes for the response
			plan.Notes = notes
		}

		// Save the updated plan
		err = s.planRepo.Update(ctx, plan)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update plan: %v", err)), nil
		}

		planJson, err := json.Marshal(plan)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal plan: %v", err)), nil
		}
		return mcp.NewToolResultText(string(planJson)), nil
	})
}

func (s *MCPGoServer) registerDeletePlanTool() {
	tool := mcp.NewTool("delete_plan",
		mcp.WithDescription("Remove a completed or cancelled feature planning plan"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Plan ID"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		err = s.planRepo.Delete(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete plan: %v", err)), nil
		}

		return mcp.NewToolResultText(`{"result":"Plan deleted"}`), nil
	})
}

func (s *MCPGoServer) registerListPlansByStatusTool() {
	tool := mcp.NewTool("list_plans_by_status",
		mcp.WithDescription("Find plans by their current status (new, inprogress, completed, cancelled)"),
		mcp.WithString("status",
			mcp.Required(),
			mcp.Description("Plan status to filter by"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		statusStr, err := request.RequireString("status")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Validate status
		status := models.PlanStatus(statusStr)
		if err := validatePlanStatus(status); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get plans with the specified status
		plans, err := s.planRepo.ListByStatus(ctx, status)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list plans by status: %v", err)), nil
		}

		plansJson, err := json.Marshal(plans)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal plans: %v", err)), nil
		}
		return mcp.NewToolResultText(string(plansJson)), nil
	})
}

// Task tools implementation

func (s *MCPGoServer) registerCreateTaskTool() {
	tool := mcp.NewTool("create_task",
		mcp.WithDescription("Create a new task as part of a feature implementation plan"),
		mcp.WithString("plan_id",
			mcp.Required(),
			mcp.Description("Plan ID this task belongs to"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Concise description of this implementation step"),
		),
		mcp.WithString("description",
			mcp.Description("Detailed explanation of what needs to be done, acceptance criteria, or implementation notes"),
		),
		mcp.WithString("status",
			mcp.Description("Current implementation status of this task (optional, defaults to 'pending')"),
			mcp.Enum("pending", "in_progress", "completed", "cancelled"),
		),
		mcp.WithString("priority",
			mcp.Description("Importance and urgency of this task in the overall feature implementation plan (optional, defaults to 'medium')"),
			mcp.Enum("low", "medium", "high"),
		),
		mcp.WithString("notes",
			mcp.Description("Initial Markdown-formatted notes for the task (optional)"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		planID, err := request.RequireString("plan_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		title, err := request.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		description := request.GetString("description", "no description provided")
		notes := request.GetString("notes", "")

		priorityStr := request.GetString("priority", string(models.TaskPriorityMedium))
		priority := models.TaskPriority(priorityStr)

		task, err := s.taskRepo.Create(ctx, planID, title, description, priority)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create task: %v", err)), nil
		}

		// If notes were provided, validate, format and update them
		if notes != "" {
			// Validate and format the markdown content
			err = markdown.Validate(notes)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid notes format: %v", err)), nil
			}

			// Sanitize and format the notes
			notes = markdown.Sanitize(notes)
			notes = markdown.Format(notes)

			err = s.taskRepo.UpdateNotes(ctx, task.ID, notes)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to set initial notes: %v", err)), nil
			}

			// Refresh task to include notes
			task, err = s.taskRepo.Get(ctx, task.ID)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to refresh task: %v", err)), nil
			}
		}

		taskJson, err := json.Marshal(task)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal task: %v", err)), nil
		}
		return mcp.NewToolResultText(string(taskJson)), nil
	})
}

func (s *MCPGoServer) registerGetTaskTool() {
	tool := mcp.NewTool("get_task",
		mcp.WithDescription("Retrieve details about a specific planned task"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		task, err := s.taskRepo.Get(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get task: %v", err)), nil
		}

		taskJson, err := json.Marshal(task)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal task: %v", err)), nil
		}
		return mcp.NewToolResultText(string(taskJson)), nil
	})
}

func (s *MCPGoServer) registerListTasksByPlanTool() {
	tool := mcp.NewTool("list_tasks_by_plan",
		mcp.WithDescription("List all tasks in a feature implementation plan"),
		mcp.WithString("plan_id",
			mcp.Required(),
			mcp.Description("Plan ID to filter tasks by"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		planID, err := request.RequireString("plan_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		tasks, err := s.taskRepo.ListByPlan(ctx, planID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list tasks by plan: %v", err)), nil
		}

		tasksJson, err := json.Marshal(tasks)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal tasks: %v", err)), nil
		}

		return mcp.NewToolResultText(string(tasksJson)), nil
	})
}

func (s *MCPGoServer) registerListTasksByStatusTool() {
	tool := mcp.NewTool("list_tasks_by_status",
		mcp.WithDescription("Find tasks by their current status (pending, in progress, completed, cancelled)"),
		mcp.WithString("status",
			mcp.Required(),
			mcp.Description("Task status to filter by"),
			mcp.Enum("pending", "in_progress", "completed", "cancelled"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		statusStr, err := request.RequireString("status")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		status := models.TaskStatus(statusStr)
		tasks, err := s.taskRepo.ListByStatus(ctx, status)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list tasks by status: %v", err)), nil
		}

		tasksJson, err := json.Marshal(tasks)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal tasks: %v", err)), nil
		}
		return mcp.NewToolResultText(string(tasksJson)), nil
	})
}

func (s *MCPGoServer) registerUpdateTaskTool() {
	tool := mcp.NewTool("update_task",
		mcp.WithDescription("Update the details, status, or priority of a planned task"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
		mcp.WithString("title",
			mcp.Description("New task title (optional)"),
		),
		mcp.WithString("description",
			mcp.Description("New task description (optional)"),
		),
		mcp.WithString("status",
			mcp.Description("New task status (optional)"),
			mcp.Enum("pending", "in_progress", "completed", "cancelled"),
		),
		mcp.WithString("priority",
			mcp.Description("New task priority (optional)"),
			mcp.Enum("low", "medium", "high"),
		),
		mcp.WithString("notes",
			mcp.Description("New Markdown-formatted notes (optional)"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get the existing task
		task, err := s.taskRepo.Get(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get task: %v", err)), nil
		}

		// Update fields if provided
		title := request.GetString("title", task.Title)
		if title != task.Title {
			task.Title = title
		}

		description := request.GetString("description", task.Description)
		if description != task.Description {
			task.Description = description
		}

		statusStr := request.GetString("status", string(task.Status))
		task.Status = models.TaskStatus(statusStr)

		priorityStr := request.GetString("priority", string(task.Priority))
		task.Priority = models.TaskPriority(priorityStr)

		// Check if notes are provided
		notes := request.GetString("notes", "")
		if notes != "" {
			// Validate and format the markdown content
			err = markdown.Validate(notes)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Invalid notes format: %v", err)), nil
			}

			// Sanitize and format the notes
			notes = markdown.Sanitize(notes)
			notes = markdown.Format(notes)

			// Update notes separately using the dedicated method
			err = s.taskRepo.UpdateNotes(ctx, id, notes)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to update notes: %v", err)), nil
			}
			// Update task.Notes for the response
			task.Notes = notes
		}

		// Save the updated task
		err = s.taskRepo.Update(ctx, task)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update task: %v", err)), nil
		}

		taskJson, err := json.Marshal(task)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal task: %v", err)), nil
		}
		return mcp.NewToolResultText(string(taskJson)), nil
	})
}

func (s *MCPGoServer) registerDeleteTaskTool() {
	tool := mcp.NewTool("delete_task",
		mcp.WithDescription("Remove a task from a feature implementation plan"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		err = s.taskRepo.Delete(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete task: %v", err)), nil
		}

		return mcp.NewToolResultText("Task deleted"), nil
	})
}

func (s *MCPGoServer) registerBulkCreateTasksTool() {
	tool := mcp.NewTool("bulk_create_tasks",
		mcp.WithDescription("Create multiple tasks at once for a feature implementation plan"),
		mcp.WithString("plan_id",
			mcp.Required(),
			mcp.Description("Plan ID these tasks belong to"),
		),
		mcp.WithString("tasks_json",
			mcp.Required(),
			mcp.Description("JSON string containing an array of task definitions, each containing title (required), description (optional), status (optional), and priority (optional)"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		planID, err := request.RequireString("plan_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Extract tasks JSON string
		tasksJSON, err := request.RequireString("tasks_json")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Unmarshal into a slice of maps
		var tasksArray []map[string]interface{}
		err = json.Unmarshal([]byte(tasksJSON), &tasksArray)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to parse tasks JSON: %v", err)), nil
		}

		// Convert tasks array to TaskCreateInput slice
		taskInputs := make([]storage.TaskCreateInput, 0, len(tasksArray))
		for _, taskMap := range tasksArray {
			// Extract title (required)
			titleRaw, ok := taskMap["title"]
			if !ok {
				return mcp.NewToolResultError("Task title is required"), nil
			}

			title, ok := titleRaw.(string)
			if !ok || title == "" {
				return mcp.NewToolResultError("Task title must be a non-empty string"), nil
			}

			// Extract optional fields
			description := ""
			if descRaw, ok := taskMap["description"]; ok {
				if desc, ok := descRaw.(string); ok {
					description = desc
				}
			}

			statusStr := ""
			if statusRaw, ok := taskMap["status"]; ok {
				if status, ok := statusRaw.(string); ok {
					statusStr = status
				}
			}

			priorityStr := ""
			if priorityRaw, ok := taskMap["priority"]; ok {
				if priority, ok := priorityRaw.(string); ok {
					priorityStr = priority
				}
			}

			// Validate status if provided
			if statusStr != "" {
				validStatus := false
				for _, s := range []string{"pending", "in_progress", "completed", "cancelled"} {
					if statusStr == s {
						validStatus = true
						break
					}
				}
				if !validStatus {
					return mcp.NewToolResultError(fmt.Sprintf("Invalid status: %s", statusStr)), nil
				}
			}

			// Validate priority if provided
			if priorityStr != "" {
				validPriority := false
				for _, p := range []string{"low", "medium", "high"} {
					if priorityStr == p {
						validPriority = true
						break
					}
				}
				if !validPriority {
					return mcp.NewToolResultError(fmt.Sprintf("Invalid priority: %s", priorityStr)), nil
				}
			}

			// Create task input
			taskInput := storage.TaskCreateInput{
				Title:       title,
				Description: description,
				Status:      models.TaskStatus(statusStr),
				Priority:    models.TaskPriority(priorityStr),
			}
			taskInputs = append(taskInputs, taskInput)
		}

		// Create tasks in bulk
		createdTasks, err := s.taskRepo.CreateBulk(ctx, planID, taskInputs)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create tasks: %v", err)), nil
		}

		// Return created tasks
		tasksJson, err := json.Marshal(createdTasks)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal tasks: %v", err)), nil
		}
		return mcp.NewToolResultText(string(tasksJson)), nil
	})
}

func (s *MCPGoServer) registerReorderTaskTool() {
	tool := mcp.NewTool("reorder_task",
		mcp.WithDescription("Change the sequence of tasks in a feature implementation plan"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
		mcp.WithNumber("new_order",
			mcp.Required(),
			mcp.Description("New order position for the task"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		newOrderFloat, err := request.RequireFloat("new_order")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		newOrder := int(newOrderFloat)

		err = s.taskRepo.ReorderTask(ctx, id, newOrder)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to reorder task: %v", err)), nil
		}

		// Get the updated task
		task, err := s.taskRepo.Get(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get updated task: %v", err)), nil
		}

		taskJson, err := json.Marshal(task)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal task: %v", err)), nil
		}
		return mcp.NewToolResultText(string(taskJson)), nil
	})
}

// registerListTasksByPlanAndStatusTool registers a tool to list tasks by both plan ID and status
func (s *MCPGoServer) registerListTasksByPlanAndStatusTool() {
	tool := mcp.NewTool("list_tasks_by_plan_and_status",
		mcp.WithDescription("Find tasks by both plan ID and status (pending, in progress, completed, cancelled)"),
		mcp.WithString("plan_id",
			mcp.Required(),
			mcp.Description("Plan ID to filter tasks by"),
		),
		mcp.WithString("status",
			mcp.Required(),
			mcp.Description("Task status to filter by"),
			mcp.Enum("pending", "in_progress", "completed", "cancelled"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		planID, err := request.RequireString("plan_id")
		if err != nil {
			return nil, err
		}

		statusStr, err := request.RequireString("status")
		if err != nil {
			return nil, err
		}

		// Convert status string to TaskStatus
		status := models.TaskStatus(statusStr)

		// Get tasks by plan ID and status
		tasks, err := s.taskRepo.ListByPlanAndStatus(ctx, planID, status)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list tasks by plan and status: %v", err)), nil
		}

		tasksJson, err := json.Marshal(tasks)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal tasks: %v", err)), nil
		}
		return mcp.NewToolResultText(string(tasksJson)), nil
	})
}

// registerListOrphanedTasksTool registers a tool to list tasks that reference non-existent plans
func (s *MCPGoServer) registerListOrphanedTasksTool() {
	tool := mcp.NewTool("list_orphaned_tasks",
		mcp.WithDescription("List all tasks that reference non-existent plans"),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Get orphaned tasks
		tasks, err := s.taskRepo.ListOrphanedTasks(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list orphaned tasks: %v", err)), nil
		}

		// Marshal tasks to JSON
		tasksJson, err := json.Marshal(tasks)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal tasks: %v", err)), nil
		}

		return mcp.NewToolResultText(string(tasksJson)), nil
	})
}

// registerUpdatePlanNotesTool registers a tool to update notes for a plan
func (s *MCPGoServer) registerUpdatePlanNotesTool() {
	tool := mcp.NewTool("update_plan_notes",
		mcp.WithDescription("Update notes for a plan"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Plan ID"),
		),
		mcp.WithString("notes",
			mcp.Required(),
			mcp.Description("Markdown-formatted notes content"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		notes, err := request.RequireString("notes")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Validate and format the markdown content
		err = markdown.Validate(notes)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid notes format: %v", err)), nil
		}

		// Sanitize and format the notes
		notes = markdown.Sanitize(notes)
		notes = markdown.Format(notes)

		// Update the notes
		err = s.planRepo.UpdateNotes(ctx, id, notes)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update plan notes: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully updated notes for plan %s", id)), nil
	})
}

// registerGetPlanNotesTool registers a tool to get notes for a plan
func (s *MCPGoServer) registerGetPlanNotesTool() {
	tool := mcp.NewTool("get_plan_notes",
		mcp.WithDescription("Retrieve the notes for a specific plan"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Plan ID"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get the notes
		notes, err := s.planRepo.GetNotes(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get plan notes: %v", err)), nil
		}

		result := map[string]string{
			"id":    id,
			"notes": notes,
		}

		resultJson, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal result: %v", err)), nil
		}
		return mcp.NewToolResultText(string(resultJson)), nil
	})
}

// registerUpdateTaskNotesTool registers a tool to update notes for a task
func (s *MCPGoServer) registerUpdateTaskNotesTool() {
	tool := mcp.NewTool("update_task_notes",
		mcp.WithDescription("Update the notes for a specific task"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
		mcp.WithString("notes",
			mcp.Required(),
			mcp.Description("Markdown-formatted notes content"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		notes, err := request.RequireString("notes")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Validate and format the markdown content
		err = markdown.Validate(notes)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid notes format: %v", err)), nil
		}

		// Sanitize and format the notes
		notes = markdown.Sanitize(notes)
		notes = markdown.Format(notes)

		// Update the notes
		err = s.taskRepo.UpdateNotes(ctx, id, notes)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update task notes: %v", err)), nil
		}

		// Get the updated task
		task, err := s.taskRepo.Get(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get updated task: %v", err)), nil
		}

		taskJson, err := json.Marshal(task)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal task: %v", err)), nil
		}
		return mcp.NewToolResultText(string(taskJson)), nil
	})
}

// registerGetTaskNotesTool registers a tool to get notes for a task
func (s *MCPGoServer) registerGetTaskNotesTool() {
	tool := mcp.NewTool("get_task_notes",
		mcp.WithDescription("Retrieve the notes for a specific task"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Task ID"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get the notes
		notes, err := s.taskRepo.GetNotes(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get task notes: %v", err)), nil
		}

		result := map[string]string{
			"id":    id,
			"notes": notes,
		}

		resultJson, err := json.Marshal(result)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal result: %v", err)), nil
		}
		return mcp.NewToolResultText(string(resultJson)), nil
	})
}

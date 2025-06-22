package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"testing"

	"github.com/jbrinkman/valkey-ai-tasks/internal/mocks"
	"github.com/jbrinkman/valkey-ai-tasks/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/internal/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock tool registry for testing
type mockToolRegistry struct {
	addToolFunc func(tool mcp.Tool, handler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error))
}

// AddTool implements the method needed for our tests
func (m *mockToolRegistry) AddTool(tool mcp.Tool, handler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)) {
	if m.addToolFunc != nil {
		m.addToolFunc(tool, handler)
	}
}

// getRandomPort returns a random available port for testing
func getRandomPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port, nil
}

// testMCPGoServer is a test-specific version of MCPGoServer that uses interfaces instead of concrete types
type testMCPGoServer struct {
	server   *mockToolRegistry
	planRepo storage.PlanRepositoryInterface
	taskRepo storage.TaskRepositoryInterface
	port     int // Added port field for test server
}

// registerListTasksByPlanAndStatusTool implements the same method as MCPGoServer for testing
func (s *testMCPGoServer) registerListTasksByPlanAndStatusTool() {
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
			return mcp.NewToolResultError(err.Error()), nil
		}

		statusStr, err := request.RequireString("status")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
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

// registerBulkCreateTasksTool implements the same method as MCPGoServer for testing
func (s *testMCPGoServer) registerBulkCreateTasksTool() {
	tool := mcp.NewTool("bulk_create_tasks",
		mcp.WithDescription("Create multiple tasks at once for a feature implementation plan"),
		mcp.WithString("plan_id", mcp.Required(), mcp.Description("Plan ID these tasks belong to")),
		mcp.WithString("tasks_json", mcp.Required(), mcp.Description("JSON string containing an array of task definitions")),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract plan_id parameter
		planID, err := request.RequireString("plan_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Extract tasks_json parameter
		tasksJSON, err := request.RequireString("tasks_json")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Parse the tasks JSON
		var taskInputs []map[string]interface{}
		if err := json.Unmarshal([]byte(tasksJSON), &taskInputs); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid tasks JSON: %v", err)), nil
		}

		// Verify the plan exists
		plan, err := s.planRepo.Get(ctx, planID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Plan not found: %v", err)), nil
		}

		// Convert to TaskCreateInput
		var createInputs []storage.TaskCreateInput
		for _, input := range taskInputs {
			title, ok := input["title"].(string)
			if !ok || title == "" {
				return mcp.NewToolResultError("Each task must have a title"), nil
			}

			description := ""
			if desc, ok := input["description"].(string); ok {
				description = desc
			}

			priority := models.TaskPriorityMedium
			if p, ok := input["priority"].(string); ok {
				switch p {
				case "low":
					priority = models.TaskPriorityLow
				case "high":
					priority = models.TaskPriorityHigh
				}
			}

			status := models.TaskStatusPending
			if s, ok := input["status"].(string); ok {
				switch s {
				case "in_progress":
					status = models.TaskStatusInProgress
				case "completed":
					status = models.TaskStatusCompleted
				case "cancelled":
					status = models.TaskStatusCancelled
				}
			}

			createInputs = append(createInputs, storage.TaskCreateInput{
				Title:       title,
				Description: description,
				Priority:    priority,
				Status:      status,
			})
		}

		// Create the tasks
		tasks, err := s.taskRepo.CreateBulk(ctx, plan.ID, createInputs)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create tasks: %v", err)), nil
		}

		// Return the created tasks as JSON
		tasksJson, err := json.Marshal(tasks)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal tasks: %v", err)), nil
		}

		return mcp.NewToolResultText(string(tasksJson)), nil
	})
}

func TestRegisterBulkCreateTasksTool(t *testing.T) {
	// Create mock repositories
	mockPlanRepo := mocks.NewMockPlanRepository()
	mockTaskRepo := mocks.NewMockTaskRepository(mockPlanRepo)

	// Create a test plan
	ctx := context.Background()
	plan, err := mockPlanRepo.Create(ctx, "app-123", "Test Plan", "Test Plan Description")
	require.NoError(t, err)
	require.NotNil(t, plan)

	// Create a test MCP server with our repositories
	server := &testMCPGoServer{
		planRepo: mockPlanRepo,
		taskRepo: mockTaskRepo,
	}

	// Create a mock tool registry to capture the registered tool
	var toolHandler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	mockRegistry := &mockToolRegistry{
		addToolFunc: func(tool mcp.Tool, handler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)) {
			toolHandler = handler
		},
	}

	// Set the mock registry as the server field
	server.server = mockRegistry

	// Register the bulk create tasks tool
	server.registerBulkCreateTasksTool()

	// Verify the tool handler was registered
	assert.NotNil(t, toolHandler)

	// Create test task data
	tasksJSON := `[
		{"title": "Task 1", "description": "Description 1", "status": "pending", "priority": "high"},
		{"title": "Task 2", "description": "", "status": "", "priority": ""},
		{"title": "Task 3", "description": "Description 3", "priority": "low"}
	]`

	// Create a mock tool request
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "bulk_create_tasks",
			Arguments: map[string]interface{}{
				"plan_id":    plan.ID,
				"tasks_json": tasksJSON,
			},
		},
	}

	// Call the tool handler
	result, err := toolHandler(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Unmarshal the result
	var tasks []*models.Task
	// Check if we got a text content result
	require.NotEmpty(t, result.Content, "Expected non-empty content")

	// Extract the text content
	var textContent string
	switch content := result.Content[0].(type) {
	case *mcp.TextContent:
		textContent = content.Text
	case mcp.TextContent:
		textContent = content.Text
	default:
		require.Fail(t, "Expected TextContent, got %T", result.Content[0])
	}

	err = json.Unmarshal([]byte(textContent), &tasks)
	require.NoError(t, err)

	// Verify we got 3 tasks back
	assert.Len(t, tasks, 3)

	// Verify task 1 properties
	assert.Equal(t, "Task 1", tasks[0].Title)
	assert.Equal(t, "Description 1", tasks[0].Description)
	assert.Equal(t, models.TaskStatusPending, tasks[0].Status)
	assert.Equal(t, models.TaskPriorityHigh, tasks[0].Priority)

	// Verify task 2 properties (with defaults)
	assert.Equal(t, "Task 2", tasks[1].Title)
	// The mock repository might set a default description
	assert.Contains(t, []string{"", "no description provided"}, tasks[1].Description)
	assert.Equal(t, models.TaskStatusPending, tasks[1].Status)
	assert.Equal(t, models.TaskPriorityMedium, tasks[1].Priority)

	// Verify task 3 properties
	assert.Equal(t, "Task 3", tasks[2].Title)
	assert.Equal(t, "Description 3", tasks[2].Description)
	assert.Equal(t, models.TaskStatusPending, tasks[2].Status)
	assert.Equal(t, models.TaskPriorityLow, tasks[2].Priority)

	// Verify the tasks were stored in the repository
	storedTasks, err := mockTaskRepo.ListByPlan(ctx, plan.ID)
	require.NoError(t, err)
	assert.Len(t, storedTasks, 3)
}

// TestRegisterListTasksByPlanAndStatusTool tests the list_tasks_by_plan_and_status MCP tool
func TestRegisterListTasksByPlanAndStatusTool(t *testing.T) {
	// Create mock repositories
	mockPlanRepo := mocks.NewMockPlanRepository()
	mockTaskRepo := mocks.NewMockTaskRepository(mockPlanRepo)

	// Create a test plan
	ctx := context.Background()
	plan, err := mockPlanRepo.Create(ctx, "app-123", "Test Plan", "Test Plan Description")
	require.NoError(t, err)
	require.NotNil(t, plan)

	// Create tasks with different statuses
	task1, err := mockTaskRepo.Create(ctx, plan.ID, "Task 1", "Description 1", models.TaskPriorityMedium)
	require.NoError(t, err)
	task1.Status = models.TaskStatusInProgress
	err = mockTaskRepo.Update(ctx, task1)
	require.NoError(t, err)

	task2, err := mockTaskRepo.Create(ctx, plan.ID, "Task 2", "Description 2", models.TaskPriorityMedium)
	require.NoError(t, err)
	task2.Status = models.TaskStatusCompleted
	err = mockTaskRepo.Update(ctx, task2)
	require.NoError(t, err)

	task3, err := mockTaskRepo.Create(ctx, plan.ID, "Task 3", "Description 3", models.TaskPriorityMedium)
	require.NoError(t, err)
	task3.Status = models.TaskStatusInProgress
	err = mockTaskRepo.Update(ctx, task3)
	require.NoError(t, err)

	// Create a test MCP server with our repositories
	server := &testMCPGoServer{
		planRepo: mockPlanRepo,
		taskRepo: mockTaskRepo,
	}

	// Create a mock tool registry to capture the registered tool
	var toolHandler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	mockRegistry := &mockToolRegistry{
		addToolFunc: func(tool mcp.Tool, handler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)) {
			toolHandler = handler
		},
	}

	// Set the mock registry as the server field
	server.server = mockRegistry

	// Register the list tasks by plan and status tool
	server.registerListTasksByPlanAndStatusTool()

	// Verify the tool handler was registered
	require.NotNil(t, toolHandler)

	// Create a mock tool request for in-progress tasks
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_tasks_by_plan_and_status",
			Arguments: map[string]interface{}{
				"plan_id": plan.ID,
				"status":  string(models.TaskStatusInProgress),
			},
		},
	}

	// Call the handler
	result, err := toolHandler(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Unmarshal the result
	var tasks []*models.Task
	// Check if we got a text content result
	require.NotEmpty(t, result.Content, "Expected non-empty content")

	// Extract the text content for in-progress tasks
	textContent1 := ""
	switch content := result.Content[0].(type) {
	case *mcp.TextContent:
		textContent1 = content.Text
	case mcp.TextContent:
		textContent1 = content.Text
	default:
		require.Fail(t, "Expected TextContent, got %T", result.Content[0])
	}

	err = json.Unmarshal([]byte(textContent1), &tasks)
	require.NoError(t, err)

	// Verify we got the expected tasks
	require.Equal(t, 2, len(tasks), "Should have found 2 in-progress tasks")
	assert.Equal(t, task1.ID, tasks[0].ID, "First task ID should match")
	assert.Equal(t, task3.ID, tasks[1].ID, "Second task ID should match")
	assert.Equal(t, models.TaskStatusInProgress, tasks[0].Status, "First task status should be in-progress")
	assert.Equal(t, models.TaskStatusInProgress, tasks[1].Status, "Second task status should be in-progress")

	// Test with completed status
	request = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_tasks_by_plan_and_status",
			Arguments: map[string]interface{}{
				"plan_id": plan.ID,
				"status":  string(models.TaskStatusCompleted),
			},
		},
	}

	result, err = toolHandler(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check if we got a text content result
	require.NotEmpty(t, result.Content, "Expected non-empty content")

	// Extract the text content for completed tasks
	textContent2 := ""
	switch content := result.Content[0].(type) {
	case *mcp.TextContent:
		textContent2 = content.Text
	case mcp.TextContent:
		textContent2 = content.Text
	default:
		require.Fail(t, "Expected TextContent, got %T", result.Content[0])
	}

	tasks = nil // Clear previous tasks
	err = json.Unmarshal([]byte(textContent2), &tasks)
	require.NoError(t, err)

	require.Equal(t, 1, len(tasks), "Should have found 1 completed task")
	assert.Equal(t, task2.ID, tasks[0].ID, "Task ID should match")
	assert.Equal(t, models.TaskStatusCompleted, tasks[0].Status, "Task status should be completed")

	// Test with invalid plan ID
	request = mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_tasks_by_plan_and_status",
			Arguments: map[string]interface{}{
				"plan_id": "invalid-plan-id",
				"status":  string(models.TaskStatusInProgress),
			},
		},
	}

	result, err = toolHandler(ctx, request)
	require.NoError(t, err)
	require.NotNil(t, result)

	// For error results, check the content
	// Extract the text content
	var textContent string
	switch content := result.Content[0].(type) {
	case *mcp.TextContent:
		textContent = content.Text
	case mcp.TextContent:
		textContent = content.Text
	default:
		require.Fail(t, "Expected TextContent, got %T", result.Content[0])
	}

	assert.Contains(t, textContent, "Failed to list tasks by plan and status")
}

func TestRegisterBulkCreateTasksToolErrors(t *testing.T) {
	// Create mock repositories
	mockPlanRepo := mocks.NewMockPlanRepository()
	mockTaskRepo := mocks.NewMockTaskRepository(mockPlanRepo)

	// Create a test MCP server with our repositories
	server := &testMCPGoServer{
		planRepo: mockPlanRepo,
		taskRepo: mockTaskRepo,
	}

	// Create a mock tool registry to capture the registered tool
	var toolHandler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

	mockRegistry := &mockToolRegistry{
		addToolFunc: func(tool mcp.Tool, handler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)) {
			toolHandler = handler
		},
	}

	// Set the mock registry as the server field
	server.server = mockRegistry

	// Register the bulk create tasks tool
	server.registerBulkCreateTasksTool()

	// Test case 1: Missing plan_id
	t.Run("MissingPlanID", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "bulk_create_tasks",
				Arguments: map[string]interface{}{
					"tasks_json": `[{"title": "Task 1"}]`,
				},
			},
		}

		result, err := toolHandler(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Extract the text content
		var textContent string
		switch content := result.Content[0].(type) {
		case *mcp.TextContent:
			textContent = content.Text
		case mcp.TextContent:
			textContent = content.Text
		default:
			require.Fail(t, "Expected TextContent, got %T", result.Content[0])
		}
		assert.Contains(t, textContent, "required argument \"plan_id\" not found")
	})

	// Test case 2: Missing tasks_json
	t.Run("MissingTasksJSON", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "bulk_create_tasks",
				Arguments: map[string]interface{}{
					"plan_id": "plan-123",
				},
			},
		}

		result, err := toolHandler(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Extract the text content
		var textContent string
		switch content := result.Content[0].(type) {
		case *mcp.TextContent:
			textContent = content.Text
		case mcp.TextContent:
			textContent = content.Text
		default:
			require.Fail(t, "Expected TextContent, got %T", result.Content[0])
		}
		assert.Contains(t, textContent, "required argument \"tasks_json\" not found")
	})

	// Test case 3: Invalid tasks_json
	t.Run("InvalidTasksJSON", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "bulk_create_tasks",
				Arguments: map[string]interface{}{
					"plan_id":    "plan-123",
					"tasks_json": "not valid json",
				},
			},
		}

		result, err := toolHandler(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Extract the text content
		var textContent string
		switch content := result.Content[0].(type) {
		case *mcp.TextContent:
			textContent = content.Text
		case mcp.TextContent:
			textContent = content.Text
		default:
			require.Fail(t, "Expected TextContent, got %T", result.Content[0])
		}
		assert.Contains(t, textContent, "Invalid tasks JSON")
	})

	// Test case 4: Plan not found
	t.Run("PlanNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "bulk_create_tasks",
				Arguments: map[string]interface{}{
					"plan_id":    "non-existent-plan",
					"tasks_json": `[{"title": "Task 1"}]`,
				},
			},
		}

		result, err := toolHandler(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Extract the text content
		var textContent string
		switch content := result.Content[0].(type) {
		case *mcp.TextContent:
			textContent = content.Text
		case mcp.TextContent:
			textContent = content.Text
		default:
			require.Fail(t, "Expected TextContent, got %T", result.Content[0])
		}
		assert.Contains(t, textContent, "Plan not found")
	})

	// Test case 5: Missing task title
	t.Run("MissingTaskTitle", func(t *testing.T) {
		// Create a test plan
		ctx := context.Background()
		plan, err := mockPlanRepo.Create(ctx, "app-456", "Test Plan", "Test Plan Description")
		require.NoError(t, err)

		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "bulk_create_tasks",
				Arguments: map[string]interface{}{
					"plan_id":    plan.ID,
					"tasks_json": `[{"description": "Missing title"}]`,
				},
			},
		}

		result, err := toolHandler(context.Background(), request)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Extract the text content
		var textContent string
		switch content := result.Content[0].(type) {
		case *mcp.TextContent:
			textContent = content.Text
		case mcp.TextContent:
			textContent = content.Text
		default:
			require.Fail(t, "Expected TextContent, got %T", result.Content[0])
		}
		assert.Contains(t, textContent, "Each task must have a title")
	})
}

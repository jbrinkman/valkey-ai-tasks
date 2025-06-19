package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jbrinkman/valkey-ai-tasks/go/internal/mocks"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/storage"
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

// testMCPGoServer is a test-specific version of MCPGoServer that uses interfaces instead of concrete types
type testMCPGoServer struct {
	server      *mockToolRegistry
	projectRepo storage.ProjectRepositoryInterface
	taskRepo    storage.TaskRepositoryInterface
}

// registerBulkCreateTasksTool implements the same method as MCPGoServer for testing
func (s *testMCPGoServer) registerBulkCreateTasksTool() {
	tool := mcp.NewTool("bulk_create_tasks",
		mcp.WithDescription("Create multiple tasks at once for a feature implementation plan"),
		mcp.WithString("project_id", mcp.Required(), mcp.Description("Project ID these tasks belong to")),
		mcp.WithString("tasks_json", mcp.Required(), mcp.Description("JSON string containing an array of task definitions")),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract project_id parameter
		projectID, err := request.RequireString("project_id")
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

		// Verify the project exists
		project, err := s.projectRepo.Get(ctx, projectID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Project not found: %v", err)), nil
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
		tasks, err := s.taskRepo.CreateBulk(ctx, project.ID, createInputs)
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
	mockProjectRepo := mocks.NewMockProjectRepository()
	mockTaskRepo := mocks.NewMockTaskRepository(mockProjectRepo)

	// Create a test project
	ctx := context.Background()
	project, err := mockProjectRepo.Create(ctx, "app-123", "Test Project", "Test Project Description")
	require.NoError(t, err)
	require.NotNil(t, project)

	// Create a test MCP server with our repositories
	server := &testMCPGoServer{
		projectRepo: mockProjectRepo,
		taskRepo:    mockTaskRepo,
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
				"project_id": project.ID,
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
	storedTasks, err := mockTaskRepo.ListByProject(ctx, project.ID)
	require.NoError(t, err)
	assert.Len(t, storedTasks, 3)
}

func TestRegisterBulkCreateTasksToolErrors(t *testing.T) {
	// Create mock repositories
	mockProjectRepo := mocks.NewMockProjectRepository()
	mockTaskRepo := mocks.NewMockTaskRepository(mockProjectRepo)

	// Create a test MCP server with our repositories
	server := &testMCPGoServer{
		projectRepo: mockProjectRepo,
		taskRepo:    mockTaskRepo,
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

	// Test case 1: Missing project_id
	t.Run("MissingProjectID", func(t *testing.T) {
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
		assert.Contains(t, textContent, "required argument \"project_id\" not found")
	})

	// Test case 2: Missing tasks_json
	t.Run("MissingTasksJSON", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "bulk_create_tasks",
				Arguments: map[string]interface{}{
					"project_id": "project-123",
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
					"project_id": "project-123",
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

	// Test case 4: Project not found
	t.Run("ProjectNotFound", func(t *testing.T) {
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "bulk_create_tasks",
				Arguments: map[string]interface{}{
					"project_id": "non-existent-project",
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
		assert.Contains(t, textContent, "Project not found")
	})

	// Test case 5: Missing task title
	t.Run("MissingTaskTitle", func(t *testing.T) {
		// Create a test project
		ctx := context.Background()
		project, err := mockProjectRepo.Create(ctx, "app-456", "Test Project", "Test Project Description")
		require.NoError(t, err)
		
		request := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name: "bulk_create_tasks",
				Arguments: map[string]interface{}{
					"project_id": project.ID,
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

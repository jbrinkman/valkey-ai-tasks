package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPGoServer wraps the mark3labs/mcp-go server implementation
type MCPGoServer struct {
	server      *server.MCPServer
	projectRepo storage.ProjectRepository
	taskRepo    storage.TaskRepository
}

// NewMCPGoServer creates a new MCP server using the mark3labs/mcp-go library
func NewMCPGoServer(projectRepo storage.ProjectRepository, taskRepo storage.TaskRepository) *MCPGoServer {
	// Create a new MCP server
	s := server.NewMCPServer(
		"Valkey Task Management",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	mcpServer := &MCPGoServer{
		server:      s,
		projectRepo: projectRepo,
		taskRepo:    taskRepo,
	}

	// Register all tools
	mcpServer.registerTools()

	return mcpServer
}

// Start starts the MCP server using the HTTP transport
func (s *MCPGoServer) Start(port int) error {
	log.Printf("Starting MCP server on port %d", port)

	// Create HTTP server with the MCP server
	httpServer := server.NewStreamableHTTPServer(s.server)

	// Start HTTP server on the specified port
	return httpServer.Start(fmt.Sprintf(":%d", port))
}

// registerTools registers all the task management tools with the MCP server
func (s *MCPGoServer) registerTools() {
	// Project tools
	s.registerCreateProjectTool()
	s.registerGetProjectTool()
	s.registerListProjectsTool()
	s.registerListProjectsByApplicationTool()
	s.registerUpdateProjectTool()
	s.registerDeleteProjectTool()

	// Task tools
	s.registerCreateTaskTool()
	s.registerGetTaskTool()
	s.registerListTasksByProjectTool()
	s.registerListTasksByStatusTool()
	s.registerUpdateTaskTool()
	s.registerDeleteTaskTool()
	s.registerReorderTaskTool()
}

// Project tools implementation

func (s *MCPGoServer) registerCreateProjectTool() {
	tool := mcp.NewTool("create_project",
		mcp.WithDescription("Create a new project"),
		mcp.WithString("application_id",
			mcp.Required(),
			mcp.Description("The application ID this project belongs to"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Project name"),
		),
		mcp.WithString("description",
			mcp.Description("Project description (optional)"),
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

		// Create the project
		project, err := s.projectRepo.Create(ctx, applicationID, name, description)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create project: %v", err)), nil
		}

		projectJson, err := json.Marshal(project)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal project: %v", err)), nil
		}
		return mcp.NewToolResultText(string(projectJson)), nil
	})
}

func (s *MCPGoServer) registerGetProjectTool() {
	tool := mcp.NewTool("get_project",
		mcp.WithDescription("Get a project by ID"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Project ID"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		project, err := s.projectRepo.Get(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get project: %v", err)), nil
		}

		projectJson, err := json.Marshal(project)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal project: %v", err)), nil
		}
		return mcp.NewToolResultText(string(projectJson)), nil
	})
}

func (s *MCPGoServer) registerListProjectsTool() {
	tool := mcp.NewTool("list_projects",
		mcp.WithDescription("List all projects"),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projects, err := s.projectRepo.List(ctx)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list projects: %v", err)), nil
		}

		projectsJson, err := json.Marshal(projects)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal projects: %v", err)), nil
		}
		return mcp.NewToolResultText(string(projectsJson)), nil
	})
}

func (s *MCPGoServer) registerListProjectsByApplicationTool() {
	tool := mcp.NewTool("list_projects_by_application",
		mcp.WithDescription("List projects by application ID"),
		mcp.WithString("application_id",
			mcp.Required(),
			mcp.Description("Application ID to filter projects by"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		applicationID, err := request.RequireString("application_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		projects, err := s.projectRepo.ListByApplication(ctx, applicationID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list projects by application: %v", err)), nil
		}

		projectsJson, err := json.Marshal(projects)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal projects: %v", err)), nil
		}
		return mcp.NewToolResultText(string(projectsJson)), nil
	})
}

func (s *MCPGoServer) registerUpdateProjectTool() {
	tool := mcp.NewTool("update_project",
		mcp.WithDescription("Update an existing project"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Project ID"),
		),
		mcp.WithString("name",
			mcp.Description("New project name (optional)"),
		),
		mcp.WithString("description",
			mcp.Description("New project description (optional)"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Get the existing project
		project, err := s.projectRepo.Get(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get project: %v", err)), nil
		}

		// Update fields if provided
		name := request.GetString("name", project.Name)
		if name != project.Name {
			project.Name = name
		}

		description := request.GetString("description", project.Description)
		if description != project.Description {
			project.Description = description
		}

		// Save the updated project
		err = s.projectRepo.Update(ctx, project)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to update project: %v", err)), nil
		}

		projectJson, err := json.Marshal(project)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal project: %v", err)), nil
		}
		return mcp.NewToolResultText(string(projectJson)), nil
	})
}

func (s *MCPGoServer) registerDeleteProjectTool() {
	tool := mcp.NewTool("delete_project",
		mcp.WithDescription("Delete a project"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Project ID"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireString("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		err = s.projectRepo.Delete(ctx, id)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to delete project: %v", err)), nil
		}

		return mcp.NewToolResultText("Project deleted"), nil
	})
}

// Task tools implementation

func (s *MCPGoServer) registerCreateTaskTool() {
	tool := mcp.NewTool("create_task",
		mcp.WithDescription("Create a new task"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("Project ID this task belongs to"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("Task title"),
		),
		mcp.WithString("description",
			mcp.Description("Task description (optional)"),
		),
		mcp.WithString("status",
			mcp.Description("Task status (optional, defaults to 'pending')"),
			mcp.Enum("pending", "in_progress", "completed", "cancelled"),
		),
		mcp.WithString("priority",
			mcp.Description("Task priority (optional, defaults to 'medium')"),
			mcp.Enum("low", "medium", "high"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectID, err := request.RequireString("project_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		title, err := request.RequireString("title")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		description := request.GetString("description", "no description provided")

		priorityStr := request.GetString("priority", string(models.TaskPriorityMedium))
		priority := models.TaskPriority(priorityStr)

		task, err := s.taskRepo.Create(ctx, projectID, title, description, priority)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create task: %v", err)), nil
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
		mcp.WithDescription("Get a task by ID"),
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

func (s *MCPGoServer) registerListTasksByProjectTool() {
	tool := mcp.NewTool("list_tasks_by_project",
		mcp.WithDescription("List tasks by project ID"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("Project ID to filter tasks by"),
		),
	)

	s.server.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectID, err := request.RequireString("project_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		tasks, err := s.taskRepo.ListByProject(ctx, projectID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list tasks by project: %v", err)), nil
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
		mcp.WithDescription("List tasks by status"),
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
		mcp.WithDescription("Update an existing task"),
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
		mcp.WithDescription("Delete a task"),
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

func (s *MCPGoServer) registerReorderTaskTool() {
	tool := mcp.NewTool("reorder_task",
		mcp.WithDescription("Change the order of a task"),
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

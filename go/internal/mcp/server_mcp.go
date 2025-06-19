package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/storage"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPGoServer wraps the mark3labs/mcp-go server implementation
type MCPGoServer struct {
	server    *server.MCPServer
	planRepo  storage.PlanRepositoryInterface
	taskRepo  storage.TaskRepositoryInterface
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

	mcpServer := &MCPGoServer{
		server:    s,
		planRepo:  planRepo,
		taskRepo:  taskRepo,
	}

	// Register all tools
	mcpServer.registerTools()

	return mcpServer
}

// Start starts the MCP server using the HTTP transport
func (s *MCPGoServer) Start(port int) error {
	log.Printf("Starting MCP server on port %d", port)

	// // Create HTTP server with the MCP server
	// httpServer := server.NewStreamableHTTPServer(s.server)

	httpServer := server.NewSSEServer(s.server)
	return httpServer.Start(fmt.Sprintf(":%d", port))
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

	// Task tools
	s.registerCreateTaskTool()
	s.registerGetTaskTool()
	s.registerListTasksByPlanTool()
	s.registerListTasksByStatusTool()
	s.registerUpdateTaskTool()
	s.registerDeleteTaskTool()
	s.registerBulkCreateTasksTool()
	s.registerReorderTaskTool()
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

		// Create the plan
		plan, err := s.planRepo.Create(ctx, applicationID, name, description)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create plan: %v", err)), nil
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
		if status != models.PlanStatusNew &&
			status != models.PlanStatusInProgress &&
			status != models.PlanStatusCompleted &&
			status != models.PlanStatusCancelled {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid status: %s", statusStr)), nil
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
		if status != models.PlanStatusNew &&
			status != models.PlanStatusInProgress &&
			status != models.PlanStatusCompleted &&
			status != models.PlanStatusCancelled {
			return mcp.NewToolResultError(fmt.Sprintf("Invalid status: %s", statusStr)), nil
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

		priorityStr := request.GetString("priority", string(models.TaskPriorityMedium))
		priority := models.TaskPriority(priorityStr)

		task, err := s.taskRepo.Create(ctx, planID, title, description, priority)
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

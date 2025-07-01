package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/jbrinkman/valkey-ai-tasks/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/internal/storage"
	"github.com/jbrinkman/valkey-ai-tasks/internal/utils/markdown"
)

// registerTaskTools registers all task-related tools with the MCP server
func (s *MCPGoServer) registerTaskTools() {
	s.registerCreateTaskTool()
	s.registerGetTaskTool()
	s.registerListTasksByPlanTool()
	s.registerListTasksByStatusTool()
	s.registerListTasksByPlanAndStatusTool()
	s.registerUpdateTaskTool()
	s.registerDeleteTaskTool()
	s.registerBulkCreateTasksTool()
	s.registerReorderTaskTool()
	s.registerListOrphanedTasksTool()
}

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
		mcp.WithString(
			"priority",
			mcp.Description(
				"Importance and urgency of this task in the overall feature implementation plan (optional, defaults to 'medium')",
			),
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
		mcp.WithString(
			"tasks_json",
			mcp.Required(),
			mcp.Description(
				"JSON string containing an array of task definitions, each containing title (required), description (optional), status (optional), and priority (optional)",
			),
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

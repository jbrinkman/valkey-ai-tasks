package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/jbrinkman/valkey-ai-tasks/internal/utils/markdown"
)

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

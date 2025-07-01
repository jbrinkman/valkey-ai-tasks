package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/jbrinkman/valkey-ai-tasks/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/internal/utils/markdown"
)

// registerPlanTools registers all plan-related tools with the MCP server
func (s *MCPGoServer) registerPlanTools() {
	s.registerCreatePlanTool()
	s.registerGetPlanTool()
	s.registerListPlansTool()
	s.registerListPlansByApplicationTool()
	s.registerUpdatePlanTool()
	s.registerDeletePlanTool()
	s.registerUpdatePlanStatusTool()
	s.registerListPlansByStatusTool()
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

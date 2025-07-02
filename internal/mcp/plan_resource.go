package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/jbrinkman/valkey-ai-tasks/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/internal/storage"
)

// Custom error types for the PlanResource
var (
	ErrInvalidURI      = errors.New("invalid resource URI")
	ErrPlanNotFound    = errors.New("plan not found")
	ErrTasksNotFound   = errors.New("tasks not found")
	ErrInvalidPlanID   = errors.New("invalid plan ID")
	ErrInvalidAppID    = errors.New("invalid application ID")
	ErrMarshalFailure  = errors.New("failed to marshal resource")
	ErrInternalStorage = errors.New("internal storage error")
)

// PlanResourceProvider implements the MCP resource provider for the PlanResource
type PlanResourceProvider struct {
	planRepo storage.PlanRepositoryInterface
	taskRepo storage.TaskRepositoryInterface
}

// NewPlanResourceProvider creates a new PlanResourceProvider
func NewPlanResourceProvider(
	planRepo storage.PlanRepositoryInterface,
	taskRepo storage.TaskRepositoryInterface,
) *PlanResourceProvider {
	return &PlanResourceProvider{
		planRepo: planRepo,
		taskRepo: taskRepo,
	}
}

// RegisterResource registers the PlanResource with the MCP server
func (p *PlanResourceProvider) RegisterResource(server *MCPGoServer) {
	// Create a resource template for accessing plan details by ID
	planTemplate := mcp.NewResourceTemplate(
		"ai-tasks://plans/{id}/full",
		"Plan Resource",
		mcp.WithTemplateDescription("Returns a complete view of a plan including its tasks and notes"),
		mcp.WithTemplateMIMEType("application/json"),
	)

	// Create a resource template for accessing all plans
	allPlansTemplate := mcp.NewResourceTemplate(
		"ai-tasks://plans/full",
		"All Plans Resource",
		mcp.WithTemplateDescription("Returns a complete view of all plans including their tasks and notes"),
		mcp.WithTemplateMIMEType("application/json"),
	)

	// Create a resource template for accessing plans by application ID
	appPlansTemplate := mcp.NewResourceTemplate(
		"ai-tasks://applications/{app_id}/plans/full",
		"Application Plans Resource",
		mcp.WithTemplateDescription(
			"Returns a complete view of all plans for a specific application including their tasks and notes",
		),
		mcp.WithTemplateMIMEType("application/json"),
	)

	// Add the templates with their handlers
	server.server.AddResourceTemplate(planTemplate, p.handleResourceRequest)
	server.server.AddResourceTemplate(allPlansTemplate, p.handleResourceRequest)
	server.server.AddResourceTemplate(appPlansTemplate, p.handleResourceRequest)
}

// handleResourceRequest handles requests for the PlanResource
func (p *PlanResourceProvider) handleResourceRequest(
	ctx context.Context,
	req mcp.ReadResourceRequest,
) ([]mcp.ResourceContents, error) {
	// Parse the URI to determine the request type
	uriInfo, err := parseResourceURI(req.Params.URI)
	if err != nil {
		// Wrap the error with more context
		return nil, fmt.Errorf("failed to parse resource URI '%s': %w", req.Params.URI, err)
	}

	// Validate that we have the required information based on request type
	switch uriInfo.requestType {
	case singlePlanRequest:
		if uriInfo.planID == "" {
			return nil, fmt.Errorf("%w: plan ID is required for single plan requests", ErrInvalidPlanID)
		}
	case appPlansRequest:
		if uriInfo.appID == "" {
			return nil, fmt.Errorf("%w: application ID is required for application plans requests", ErrInvalidAppID)
		}
	}

	// Handle different URI patterns
	switch uriInfo.requestType {
	case singlePlanRequest:
		return p.handleSinglePlanRequest(ctx, uriInfo.planID)
	case allPlansRequest:
		return p.handleAllPlansRequest(ctx)
	case appPlansRequest:
		return p.handleAppPlansRequest(ctx, uriInfo.appID)
	default:
		return nil, fmt.Errorf("%w: unsupported request type for URI: %s", ErrInvalidURI, req.Params.URI)
	}
}

// handleSinglePlanRequest handles requests for a single plan
func (p *PlanResourceProvider) handleSinglePlanRequest(ctx context.Context, planID string) ([]mcp.ResourceContents, error) {
	// Validate plan ID
	if strings.TrimSpace(planID) == "" {
		return nil, fmt.Errorf("%w: empty plan ID", ErrInvalidPlanID)
	}

	// Get the plan
	plan, err := p.planRepo.Get(ctx, planID)
	if err != nil {
		// Check for not found case by examining error message
		if strings.Contains(err.Error(), "plan not found") {
			return nil, fmt.Errorf("%w: plan with ID '%s' does not exist", ErrPlanNotFound, planID)
		}
		// Handle other errors
		return nil, fmt.Errorf("%w: failed to get plan with ID '%s': %v", ErrInternalStorage, planID, err)
	}

	// Validate plan is not nil
	if plan == nil {
		return nil, fmt.Errorf("%w: plan with ID '%s' returned nil", ErrPlanNotFound, planID)
	}

	// Get tasks for the plan
	tasks, err := p.taskRepo.ListByPlan(ctx, planID)
	if err != nil {
		// Handle task retrieval errors
		return nil, fmt.Errorf("%w: failed to get tasks for plan '%s': %v", ErrInternalStorage, planID, err)
	}

	// Note: Empty tasks list is valid, so we don't check for nil or empty

	// Create the plan resource
	planResource := models.NewPlanResource(plan, tasks)

	// Convert to JSON
	jsonData, err := json.MarshalIndent(planResource, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal plan resource for plan '%s': %v", ErrMarshalFailure, planID, err)
	}

	// Return the resource contents
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      fmt.Sprintf("ai-tasks://plans/%s/full", planID),
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

// handleAllPlansRequest handles requests for all plans
func (p *PlanResourceProvider) handleAllPlansRequest(ctx context.Context) ([]mcp.ResourceContents, error) {
	// Get all plans
	plans, err := p.planRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to list plans: %v", ErrInternalStorage, err)
	}

	// Handle empty plans list
	if len(plans) == 0 {
		// Return empty array instead of error
		emptyJSON := "[]"
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      "ai-tasks://plans/full",
				MIMEType: "application/json",
				Text:     emptyJSON,
			},
		}, nil
	}

	// Create a list of plan resources
	planResources := make([]*models.PlanResource, 0, len(plans))
	for _, plan := range plans {
		// Get tasks for the plan
		tasks, err := p.taskRepo.ListByPlan(ctx, plan.ID)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to get tasks for plan '%s': %v", ErrInternalStorage, plan.ID, err)
		}

		// Create the plan resource
		planResource := models.NewPlanResource(plan, tasks)
		planResources = append(planResources, planResource)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(planResources, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal multiple plan resources: %v", ErrMarshalFailure, err)
	}

	// Return the resource contents
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "ai-tasks://plans/full",
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

// handleAppPlansRequest handles requests for plans by application ID
func (p *PlanResourceProvider) handleAppPlansRequest(ctx context.Context, appID string) ([]mcp.ResourceContents, error) {
	// Validate application ID
	if strings.TrimSpace(appID) == "" {
		return nil, fmt.Errorf("%w: empty application ID", ErrInvalidAppID)
	}

	// Get plans for the application
	plans, err := p.planRepo.ListByApplication(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get plans for application '%s': %v", ErrInternalStorage, appID, err)
	}

	// Handle empty plans list
	if len(plans) == 0 {
		// Return empty array instead of error
		emptyJSON := "[]"
		return []mcp.ResourceContents{
			mcp.TextResourceContents{
				URI:      fmt.Sprintf("ai-tasks://applications/%s/plans/full", appID),
				MIMEType: "application/json",
				Text:     emptyJSON,
			},
		}, nil
	}

	// Create a list of plan resources
	planResources := make([]*models.PlanResource, 0, len(plans))
	for _, plan := range plans {
		// Get tasks for the plan
		tasks, err := p.taskRepo.ListByPlan(ctx, plan.ID)
		if err != nil {
			return nil, fmt.Errorf(
				"%w: failed to get tasks for plan '%s' in application '%s': %v",
				ErrInternalStorage,
				plan.ID,
				appID,
				err,
			)
		}

		// Create the plan resource
		planResource := models.NewPlanResource(plan, tasks)
		planResources = append(planResources, planResource)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(planResources, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal application plan resources: %v", ErrMarshalFailure, err)
	}

	// Return the resource contents
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      fmt.Sprintf("ai-tasks://applications/%s/plans/full", appID),
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

// requestType represents the type of resource request
type requestType int

const (
	unknownRequest requestType = iota
	singlePlanRequest
	allPlansRequest
	appPlansRequest
)

// uriInfo contains information parsed from a resource URI
type uriInfo struct {
	requestType requestType
	planID      string
	appID       string
}

// URI patterns for resource parsing
var (
	// Pattern for single plan: ai-tasks://plans/{id}/full
	singlePlanPattern = regexp.MustCompile(`ai-tasks://plans/([^/]+)/full$`)

	// Pattern for all plans: ai-tasks://plans/full
	allPlansPattern = regexp.MustCompile(`ai-tasks://plans/full$`)

	// Pattern for application plans: ai-tasks://applications/{app_id}/plans/full
	appPlansPattern = regexp.MustCompile(`ai-tasks://applications/([^/]+)/plans/full$`)
)

// parseResourceURI parses a resource URI and extracts relevant information
func parseResourceURI(uri string) (*uriInfo, error) {
	// Check for single plan pattern
	if matches := singlePlanPattern.FindStringSubmatch(uri); len(matches) == 2 {
		return &uriInfo{
			requestType: singlePlanRequest,
			planID:      matches[1],
		}, nil
	}

	// Check for all plans pattern
	if allPlansPattern.MatchString(uri) {
		return &uriInfo{
			requestType: allPlansRequest,
		}, nil
	}

	// Check for application plans pattern
	if matches := appPlansPattern.FindStringSubmatch(uri); len(matches) == 2 {
		return &uriInfo{
			requestType: appPlansRequest,
			appID:       matches[1],
		}, nil
	}

	// Provide detailed error message for unsupported URI format
	return nil, fmt.Errorf(
		"%w: '%s' does not match any supported pattern. Expected formats: 'ai-tasks://plans/{id}/full', 'ai-tasks://plans/full', or 'ai-tasks://applications/{app_id}/plans/full'",
		ErrInvalidURI,
		uri,
	)
}

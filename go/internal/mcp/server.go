package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/storage"
)

// MCPServer implements the Model Context Protocol server for task management
type MCPServer struct {
	planRepo *storage.PlanRepository
	taskRepo *storage.TaskRepository
}

// NewMCPServer creates a new MCP server with the given repositories
func NewMCPServer(planRepo *storage.PlanRepository, taskRepo *storage.TaskRepository) *MCPServer {
	return &MCPServer{
		planRepo: planRepo,
		taskRepo: taskRepo,
	}
}

// ServeHTTP implements the http.Handler interface for the MCP server
func (s *MCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set common headers
	w.Header().Set("Content-Type", "application/json")

	// Handle OPTIONS requests for CORS
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Set CORS headers for all responses
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Parse the path to determine the operation
	path := strings.TrimPrefix(r.URL.Path, "/")
	pathParts := strings.Split(path, "/")

	log.Printf("Url: %s", r.RequestURI)
	log.Printf("Method: %s", r.Method)
	log.Printf("Path parts: %v", pathParts)
	if len(pathParts) == 0 {
		handleError(w, fmt.Errorf("invalid path"), http.StatusBadRequest)
		return
	}

	// Handle MCP protocol endpoints
	if pathParts[0] == "mcp" {
		// Check if this is an SSE request
		if len(pathParts) >= 2 && pathParts[1] == "sse" {
			s.handleSSE(w, r)
			return
		}

		// Handle other MCP requests
		s.handleMCPRequest(w, r, pathParts[1:])
		return
	}

	// If we get here, it's an unknown endpoint
	handleError(w, fmt.Errorf("unknown endpoint: %s", pathParts[0]), http.StatusNotFound)
}

// handleMCPRequest processes MCP protocol requests
func (s *MCPServer) handleMCPRequest(w http.ResponseWriter, r *http.Request, pathParts []string) {
	// If it's a GET request to the root /mcp endpoint, return the list of functions
	if len(pathParts) == 0 && r.Method == http.MethodGet {
		s.handleListFunctions(w, r)
		return
	}

	// Otherwise, handle specific paths
	if len(pathParts) == 0 {
		handleError(w, fmt.Errorf("invalid MCP path"), http.StatusBadRequest)
		return
	}

	switch pathParts[0] {
	case "list_functions":
		s.handleListFunctions(w, r)
	case "invoke":
		s.handleInvoke(w, r, pathParts[1:])
	default:
		handleError(w, fmt.Errorf("unknown MCP operation: %s", pathParts[0]), http.StatusNotFound)
	}
}

// MCPFunction represents a function that can be called via the MCP protocol
type MCPFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// MCPFunctionList represents the list of available functions
type MCPFunctionList struct {
	Functions []MCPFunction `json:"functions"`
}

// handleListFunctions returns the list of available functions
func (s *MCPServer) handleListFunctions(w http.ResponseWriter, r *http.Request) {
	functions := MCPFunctionList{
		Functions: []MCPFunction{
			{
				Name:        "create_plan",
				Description: "Create a new plan for planning and organizing a feature or initiative",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"application_id": map[string]interface{}{
							"type":        "string",
							"description": "Application identifier to associate with this plan",
						},
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Plan name",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Plan description",
						},
					},
					"required": []string{"application_id", "name"},
				},
			},
			{
				Name:        "get_plan",
				Description: "Get a plan by ID",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "string",
							"description": "Project ID",
						},
					},
					"required": []string{"id"},
				},
			},
			{
				Name:        "list_plans",
				Description: "List all plans",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
			{
				Name:        "list_plans_by_application",
				Description: "List all plans for a specific application",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"application_id": map[string]interface{}{
							"type":        "string",
							"description": "Application identifier to filter plans by",
						},
					},
					"required": []string{"application_id"},
				},
			},
			{
				Name:        "update_plan",
				Description: "Update an existing plan",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "string",
							"description": "Project ID",
						},
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Project name",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Project description",
						},
					},
					"required": []string{"id"},
				},
			},
			{
				Name:        "delete_plan",
				Description: "Delete a plan by ID",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "string",
							"description": "Project ID",
						},
					},
					"required": []string{"id"},
				},
			},
			{
				Name:        "update_plan_status",
				Description: "Update the status of a plan",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "string",
							"description": "Plan ID",
						},
						"status": map[string]interface{}{
							"type":        "string",
							"description": "Plan status (new, inprogress, completed, cancelled)",
							"enum":        []string{"new", "inprogress", "completed", "cancelled"},
						},
					},
					"required": []string{"id", "status"},
				},
			},
			{
				Name:        "list_plans_by_status",
				Description: "List all plans with a specific status",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"status": map[string]interface{}{
							"type":        "string",
							"description": "Plan status (new, inprogress, completed, cancelled)",
							"enum":        []string{"new", "inprogress", "completed", "cancelled"},
						},
					},
					"required": []string{"status"},
				},
			},
			{
				Name:        "create_task",
				Description: "Create a new task as part of a feature implementation plan",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"plan_id": map[string]interface{}{
							"type":        "string",
							"description": "Plan ID this task belongs to",
						},
						"title": map[string]interface{}{
							"type":        "string",
							"description": "Task title",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Task description",
						},
						"priority": map[string]interface{}{
							"type":        "string",
							"description": "Task priority (low, medium, high)",
							"enum":        []string{"low", "medium", "high"},
						},
					},
					"required": []string{"plan_id", "title"},
				},
			},
			{
				Name:        "get_task",
				Description: "Get a task by ID",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "string",
							"description": "Task ID",
						},
					},
					"required": []string{"id"},
				},
			},
			{
				Name:        "list_tasks_by_plan",
				Description: "List all tasks in a feature implementation plan",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"plan_id": map[string]interface{}{
							"type":        "string",
							"description": "Plan ID to filter tasks by",
						},
					},
					"required": []string{"plan_id"},
				},
			},
			{
				Name:        "list_tasks_by_status",
				Description: "List all tasks with a specific status",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"status": map[string]interface{}{
							"type":        "string",
							"description": "Task status (pending, in_progress, completed, cancelled)",
							"enum":        []string{"pending", "in_progress", "completed", "cancelled"},
						},
					},
					"required": []string{"status"},
				},
			},
			{
				Name:        "update_task",
				Description: "Update an existing task",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "string",
							"description": "Task ID",
						},
						"title": map[string]interface{}{
							"type":        "string",
							"description": "Task title",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Task description",
						},
						"status": map[string]interface{}{
							"type":        "string",
							"description": "Task status (pending, in_progress, completed, cancelled)",
							"enum":        []string{"pending", "in_progress", "completed", "cancelled"},
						},
						"priority": map[string]interface{}{
							"type":        "string",
							"description": "Task priority (low, medium, high)",
							"enum":        []string{"low", "medium", "high"},
						},
						"plan_id": map[string]interface{}{
							"type":        "string",
							"description": "Plan ID (if moving to another plan)",
						},
					},
					"required": []string{"id"},
				},
			},
			{
				Name:        "delete_task",
				Description: "Delete a task by ID",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "string",
							"description": "Task ID",
						},
					},
					"required": []string{"id"},
				},
			},
			{
				Name:        "reorder_task",
				Description: "Change the order of a task within its plan",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "string",
							"description": "Task ID",
						},
						"new_order": map[string]interface{}{
							"type":        "integer",
							"description": "New position in the plan's task list (0-based)",
						},
					},
					"required": []string{"id", "new_order"},
				},
			},
		},
	}

	json.NewEncoder(w).Encode(functions)
}

// handleInvoke processes function invocation requests
func (s *MCPServer) handleInvoke(w http.ResponseWriter, r *http.Request, pathParts []string) {
	if len(pathParts) == 0 {
		handleError(w, fmt.Errorf("function name required"), http.StatusBadRequest)
		return
	}

	functionName := pathParts[0]

	// Parse the request body
	var params map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		handleError(w, fmt.Errorf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Create a context for the operation
	ctx := context.Background()

	// Handle the function invocation
	var result interface{}
	var err error

	switch functionName {
	case "create_plan":
		result, err = s.createPlan(ctx, params)
	case "get_plan":
		result, err = s.getPlan(ctx, params)
	case "list_plans":
		result, err = s.listPlans(ctx)
	case "list_plans_by_application":
		result, err = s.listPlansByApplication(ctx, params)
	case "update_plan":
		result, err = s.updatePlan(ctx, params)
	case "delete_plan":
		result, err = s.deletePlan(ctx, params)
	case "update_plan_status":
		result, err = s.updatePlanStatus(ctx, params)
	case "list_plans_by_status":
		result, err = s.listPlansByStatus(ctx, params)
	case "create_task":
		result, err = s.createTask(ctx, params)
	case "get_task":
		result, err = s.getTask(ctx, params)
	case "list_tasks_by_plan":
		result, err = s.listTasksByPlan(ctx, params)
	case "list_tasks_by_status":
		result, err = s.listTasksByStatus(ctx, params)
	case "update_task":
		result, err = s.updateTask(ctx, params)
	case "delete_task":
		result, err = s.deleteTask(ctx, params)
	case "reorder_task":
		result, err = s.reorderTask(ctx, params)
	default:
		handleError(w, fmt.Errorf("unknown function: %s", functionName), http.StatusNotFound)
		return
	}

	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}

	// Return the result
	json.NewEncoder(w).Encode(result)
}

// Helper function to handle errors
func handleError(w http.ResponseWriter, err error, statusCode int) {
	log.Printf("Error: %v", err)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
}

// Function implementations for plan operations
func (s *MCPServer) createPlan(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	applicationID, ok := params["application_id"].(string)
	if !ok {
		return nil, fmt.Errorf("application_id is required and must be a string")
	}

	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name is required and must be a string")
	}

	description, _ := params["description"].(string) // Optional

	plan, err := s.planRepo.Create(ctx, applicationID, name, description)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (s *MCPServer) getPlan(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	id, ok := params["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required and must be a string")
	}

	plan, err := s.planRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (s *MCPServer) listPlans(ctx context.Context) (interface{}, error) {
	plans, err := s.planRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}
	return plans, nil
}

func (s *MCPServer) listPlansByApplication(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	applicationID, ok := params["application_id"].(string)
	if !ok {
		return nil, fmt.Errorf("application_id is required and must be a string")
	}

	plans, err := s.planRepo.ListByApplication(ctx, applicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to list plans by application: %w", err)
	}

	return plans, nil
}

func (s *MCPServer) updatePlan(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	id, ok := params["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required and must be a string")
	}

	// Get the existing plan
	plan, err := s.planRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if name, ok := params["name"].(string); ok {
		plan.Name = name
	}

	if description, ok := params["description"].(string); ok {
		plan.Description = description
	}

	// Update the plan
	err = s.planRepo.Update(ctx, plan)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (s *MCPServer) deletePlan(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	id, ok := params["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required and must be a string")
	}

	err := s.planRepo.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	return map[string]string{"result": "Plan deleted"}, nil
}

func (s *MCPServer) updatePlanStatus(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	id, ok := params["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required and must be a string")
	}

	statusStr, ok := params["status"].(string)
	if !ok {
		return nil, fmt.Errorf("status is required and must be a string")
	}

	// Validate status
	status := models.PlanStatus(statusStr)
	if status != models.PlanStatusNew &&
		status != models.PlanStatusInProgress &&
		status != models.PlanStatusCompleted &&
		status != models.PlanStatusCancelled {
		return nil, fmt.Errorf("invalid status: %s", statusStr)
	}

	// Get the existing plan
	plan, err := s.planRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update status
	plan.Status = status
	plan.UpdatedAt = time.Now()

	// Save the updated plan
	err = s.planRepo.Update(ctx, plan)
	if err != nil {
		return nil, err
	}

	return plan, nil
}

func (s *MCPServer) listPlansByStatus(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	statusStr, ok := params["status"].(string)
	if !ok {
		return nil, fmt.Errorf("status is required and must be a string")
	}

	// Validate status
	status := models.PlanStatus(statusStr)
	if status != models.PlanStatusNew &&
		status != models.PlanStatusInProgress &&
		status != models.PlanStatusCompleted &&
		status != models.PlanStatusCancelled {
		return nil, fmt.Errorf("invalid status: %s", statusStr)
	}

	// Get plans by status
	plans, err := s.planRepo.ListByStatus(ctx, status)
	if err != nil {
		return nil, err
	}

	return plans, nil
}

// Function implementations for task operations
func (s *MCPServer) createTask(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	planID, ok := params["plan_id"].(string)
	if !ok {
		return nil, fmt.Errorf("plan_id is required and must be a string")
	}

	title, ok := params["title"].(string)
	if !ok {
		return nil, fmt.Errorf("title is required and must be a string")
	}

	description, _ := params["description"].(string) // Optional

	priorityStr, ok := params["priority"].(string)
	if !ok {
		priorityStr = string(models.TaskPriorityMedium) // Default priority
	}

	priority := models.TaskPriority(priorityStr)
	if priority != models.TaskPriorityLow &&
		priority != models.TaskPriorityMedium &&
		priority != models.TaskPriorityHigh {
		return nil, fmt.Errorf("invalid priority: %s", priorityStr)
	}

	task, err := s.taskRepo.Create(ctx, planID, title, description, priority)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *MCPServer) getTask(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	id, ok := params["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required and must be a string")
	}

	task, err := s.taskRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *MCPServer) listTasksByPlan(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	planID, ok := params["plan_id"].(string)
	if !ok {
		return nil, fmt.Errorf("plan_id is required and must be a string")
	}

	tasks, err := s.taskRepo.ListByPlan(ctx, planID)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *MCPServer) listTasksByStatus(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	statusStr, ok := params["status"].(string)
	if !ok {
		return nil, fmt.Errorf("status is required and must be a string")
	}

	status := models.TaskStatus(statusStr)
	if status != models.TaskStatusPending &&
		status != models.TaskStatusInProgress &&
		status != models.TaskStatusCompleted &&
		status != models.TaskStatusCancelled {
		return nil, fmt.Errorf("invalid status: %s", statusStr)
	}

	tasks, err := s.taskRepo.ListByStatus(ctx, status)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *MCPServer) updateTask(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	id, ok := params["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required and must be a string")
	}

	// Get the existing task
	task, err := s.taskRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if title, ok := params["title"].(string); ok {
		task.Title = title
	}

	if description, ok := params["description"].(string); ok {
		task.Description = description
	}

	if statusStr, ok := params["status"].(string); ok {
		status := models.TaskStatus(statusStr)
		if status != models.TaskStatusPending &&
			status != models.TaskStatusInProgress &&
			status != models.TaskStatusCompleted &&
			status != models.TaskStatusCancelled {
			return nil, fmt.Errorf("invalid status: %s", statusStr)
		}
		task.Status = status
	}

	if priorityStr, ok := params["priority"].(string); ok {
		priority := models.TaskPriority(priorityStr)
		if priority != models.TaskPriorityLow &&
			priority != models.TaskPriorityMedium &&
			priority != models.TaskPriorityHigh {
			return nil, fmt.Errorf("invalid priority: %s", priorityStr)
		}
		task.Priority = priority
	}

	if planID, ok := params["plan_id"].(string); ok && planID != task.PlanID {
		// Check if the new plan exists
		_, err := s.planRepo.Get(ctx, planID)
		if err != nil {
			return nil, fmt.Errorf("new plan not found: %s", planID)
		}
		task.PlanID = planID
	}

	// Update the task
	err = s.taskRepo.Update(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *MCPServer) deleteTask(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	id, ok := params["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required and must be a string")
	}

	err := s.taskRepo.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	return map[string]string{"status": "success", "message": "Task deleted"}, nil
}

func (s *MCPServer) reorderTask(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	id, ok := params["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required and must be a string")
	}

	newOrderFloat, ok := params["new_order"].(float64)
	if !ok {
		return nil, fmt.Errorf("new_order is required and must be a number")
	}
	newOrder := int(newOrderFloat)

	err := s.taskRepo.ReorderTask(ctx, id, newOrder)
	if err != nil {
		return nil, err
	}

	// Get the updated task
	task, err := s.taskRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// handleSSE handles Server-Sent Events connections for the MCP protocol
func (s *MCPServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	log.Printf("SSE connection attempt from %s", r.RemoteAddr)

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel for client disconnection detection
	clientGone := r.Context().Done()

	// Send initial connection established message - this is critical
	// Format must be exactly as expected by MCP clients
	fmt.Fprintf(w, "data: {\"type\":\"connection_established\"}\n\n")
	w.(http.Flusher).Flush()

	log.Printf("SSE connection established with %s", r.RemoteAddr)

	// If this is a POST request, handle function invocation
	if r.Method == http.MethodPost {
		var requestData map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			log.Printf("Error parsing SSE request body: %v", err)
			sendSSEError(w, "invalid_request", "Failed to parse request body", http.StatusBadRequest)
			return
		}

		// Extract function name and parameters
		functionName, ok := requestData["name"].(string)
		if !ok {
			log.Printf("Missing function name in SSE request")
			sendSSEError(w, "invalid_request", "Missing or invalid function name", http.StatusBadRequest)
			return
		}

		log.Printf("SSE function call: %s", functionName)

		// Get parameters (if any)
		var params map[string]interface{}
		if p, ok := requestData["parameters"].(map[string]interface{}); ok {
			params = p
		} else {
			params = make(map[string]interface{})
		}

		// Generate a request ID
		requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())

		// Execute the appropriate function based on the name
		var result interface{}
		var err error

		switch functionName {
		case "create_plan":
			result, err = s.createPlan(r.Context(), params)
		case "get_plan":
			result, err = s.getPlan(r.Context(), params)
		case "list_plans":
			result, err = s.listPlans(r.Context())
		case "list_plans_by_application":
			result, err = s.listPlansByApplication(r.Context(), params)
		case "update_plan":
			result, err = s.updatePlan(r.Context(), params)
		case "delete_plan":
			result, err = s.deletePlan(r.Context(), params)
		case "create_task":
			result, err = s.createTask(r.Context(), params)
		case "get_task":
			result, err = s.getTask(r.Context(), params)
		case "list_tasks_by_plan":
			result, err = s.listTasksByPlan(r.Context(), params)
		case "list_tasks_by_status":
			result, err = s.listTasksByStatus(r.Context(), params)
		case "update_task":
			result, err = s.updateTask(r.Context(), params)
		case "delete_task":
			result, err = s.deleteTask(r.Context(), params)
		case "reorder_task":
			result, err = s.reorderTask(r.Context(), params)
		default:
			log.Printf("Unknown function in SSE request: %s", functionName)
			sendSSEError(w, "unknown_function", fmt.Sprintf("Unknown function: %s", functionName), http.StatusBadRequest)
			return
		}

		if err != nil {
			log.Printf("Error executing function %s: %v", functionName, err)
			sendSSEError(w, "execution_error", err.Error(), http.StatusInternalServerError)
			return
		}

		// Send the result
		responseData := map[string]interface{}{
			"id":     requestID,
			"result": result,
		}
		responseJSON, _ := json.Marshal(responseData)

		fmt.Fprintf(w, "data: %s\n\n", responseJSON)
		w.(http.Flusher).Flush()
		log.Printf("SSE function result sent: %s", functionName)
	}

	// Keep the connection alive with periodic heartbeats
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	heartbeatCount := 0
	for {
		select {
		case <-ticker.C:
			// Send heartbeat
			heartbeatCount++
			fmt.Fprintf(w, "data: {\"type\":\"heartbeat\",\"count\":%d}\n\n", heartbeatCount)
			w.(http.Flusher).Flush()
			log.Printf("SSE heartbeat sent (%d)", heartbeatCount)
		case <-clientGone:
			// Client disconnected
			log.Printf("SSE client disconnected: %s", r.RemoteAddr)
			return
		}
	}
}

// sendSSEError sends an error event over the SSE connection
func sendSSEError(w http.ResponseWriter, errorType, message string, statusCode int) {
	errorJSON, _ := json.Marshal(map[string]interface{}{
		"error": map[string]interface{}{
			"type":    errorType,
			"message": message,
			"code":    statusCode,
		},
	})
	fmt.Fprintf(w, "data: %s\n\n", errorJSON)
	w.(http.Flusher).Flush()
	log.Printf("SSE error sent: %s - %s", errorType, message)
}

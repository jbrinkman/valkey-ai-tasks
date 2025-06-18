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
	projectRepo *storage.ProjectRepository
	taskRepo    *storage.TaskRepository
}

// NewMCPServer creates a new MCP server with the given repositories
func NewMCPServer(projectRepo *storage.ProjectRepository, taskRepo *storage.TaskRepository) *MCPServer {
	return &MCPServer{
		projectRepo: projectRepo,
		taskRepo:    taskRepo,
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
				Name:        "create_project",
				Description: "Create a new project",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"application_id": map[string]interface{}{
							"type":        "string",
							"description": "Application identifier to associate with this project",
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
					"required": []string{"application_id", "name"},
				},
			},
			{
				Name:        "get_project",
				Description: "Get a project by ID",
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
				Name:        "list_projects",
				Description: "List all projects",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
			{
				Name:        "list_projects_by_application",
				Description: "List all projects for a specific application",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"application_id": map[string]interface{}{
							"type":        "string",
							"description": "Application identifier to filter projects by",
						},
					},
					"required": []string{"application_id"},
				},
			},
			{
				Name:        "update_project",
				Description: "Update an existing project",
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
				Name:        "delete_project",
				Description: "Delete a project by ID",
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
				Name:        "create_task",
				Description: "Create a new task in a project",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"project_id": map[string]interface{}{
							"type":        "string",
							"description": "Project ID",
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
					"required": []string{"project_id", "title"},
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
				Name:        "list_tasks_by_project",
				Description: "List all tasks in a project",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"project_id": map[string]interface{}{
							"type":        "string",
							"description": "Project ID",
						},
					},
					"required": []string{"project_id"},
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
						"project_id": map[string]interface{}{
							"type":        "string",
							"description": "Project ID (if moving to another project)",
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
				Description: "Change the order of a task within its project",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "string",
							"description": "Task ID",
						},
						"new_order": map[string]interface{}{
							"type":        "integer",
							"description": "New position in the project's task list (0-based)",
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
	case "create_project":
		result, err = s.createProject(ctx, params)
	case "get_project":
		result, err = s.getProject(ctx, params)
	case "list_projects":
		result, err = s.listProjects(ctx)
	case "list_projects_by_application":
		result, err = s.listProjectsByApplication(ctx, params)
	case "update_project":
		result, err = s.updateProject(ctx, params)
	case "delete_project":
		result, err = s.deleteProject(ctx, params)
	case "create_task":
		result, err = s.createTask(ctx, params)
	case "get_task":
		result, err = s.getTask(ctx, params)
	case "list_tasks_by_project":
		result, err = s.listTasksByProject(ctx, params)
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

// Function implementations for project operations
func (s *MCPServer) createProject(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	applicationID, ok := params["application_id"].(string)
	if !ok {
		return nil, fmt.Errorf("application_id is required and must be a string")
	}

	name, ok := params["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name is required and must be a string")
	}

	description, _ := params["description"].(string) // Optional

	project, err := s.projectRepo.Create(ctx, applicationID, name, description)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *MCPServer) getProject(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	id, ok := params["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required and must be a string")
	}

	project, err := s.projectRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *MCPServer) listProjects(ctx context.Context) (interface{}, error) {
	projects, err := s.projectRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (s *MCPServer) listProjectsByApplication(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	applicationID, ok := params["application_id"].(string)
	if !ok {
		return nil, fmt.Errorf("application_id is required and must be a string")
	}

	projects, err := s.projectRepo.ListByApplication(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (s *MCPServer) updateProject(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	id, ok := params["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required and must be a string")
	}

	// Get the existing project
	project, err := s.projectRepo.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if name, ok := params["name"].(string); ok {
		project.Name = name
	}

	if description, ok := params["description"].(string); ok {
		project.Description = description
	}

	// Update the project
	err = s.projectRepo.Update(ctx, project)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *MCPServer) deleteProject(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	id, ok := params["id"].(string)
	if !ok {
		return nil, fmt.Errorf("id is required and must be a string")
	}

	err := s.projectRepo.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	return map[string]string{"status": "success", "message": "Project deleted"}, nil
}

// Function implementations for task operations
func (s *MCPServer) createTask(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	projectID, ok := params["project_id"].(string)
	if !ok {
		return nil, fmt.Errorf("project_id is required and must be a string")
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

	task, err := s.taskRepo.Create(ctx, projectID, title, description, priority)
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

func (s *MCPServer) listTasksByProject(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	projectID, ok := params["project_id"].(string)
	if !ok {
		return nil, fmt.Errorf("project_id is required and must be a string")
	}

	tasks, err := s.taskRepo.ListByProject(ctx, projectID)
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

	if projectID, ok := params["project_id"].(string); ok && projectID != task.ProjectID {
		// Check if the new project exists
		_, err := s.projectRepo.Get(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("new project not found: %s", projectID)
		}
		task.ProjectID = projectID
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
		case "create_project":
			result, err = s.createProject(r.Context(), params)
		case "get_project":
			result, err = s.getProject(r.Context(), params)
		case "list_projects":
			result, err = s.listProjects(r.Context())
		case "list_projects_by_application":
			result, err = s.listProjectsByApplication(r.Context(), params)
		case "update_project":
			result, err = s.updateProject(r.Context(), params)
		case "delete_project":
			result, err = s.deleteProject(r.Context(), params)
		case "create_task":
			result, err = s.createTask(r.Context(), params)
		case "get_task":
			result, err = s.getTask(r.Context(), params)
		case "list_tasks_by_project":
			result, err = s.listTasksByProject(r.Context(), params)
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

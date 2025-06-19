package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/storage"
	"github.com/jbrinkman/valkey-ai-tasks/go/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPBulkCreateTasks tests the MCP bulk_create_tasks tool
func TestMCPBulkCreateTasks(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set up test context
	ctx := context.Background()
	req := require.New(t)

	// Start a Valkey container
	container, err := utils.StartValkeyContainer(ctx, t)
	req.NoError(err, "Failed to start Valkey container")
	defer utils.StopValkeyContainer(ctx, t, container)

	// Extract host and port from container URI
	uri := container.URI
	host := "localhost" // Default to localhost
	port := "6379"      // Default port

	// Parse the URI to extract host and port if needed
	if len(uri) > 8 { // "redis://" is 8 chars
		hostPort := uri[8:] // Remove "redis://" prefix
		if hostPort != "" {
			parts := utils.ParseHostPort(hostPort)
			if len(parts) == 2 {
				host = parts[0]
				port = parts[1]
			}
		}
	}

	// Create Valkey client
	valkeyClient, err := storage.NewValkeyClient(host, utils.ParseInt(port), "", "")
	req.NoError(err, "Failed to create Valkey client")
	defer valkeyClient.Close()

	// Create repositories
	planRepo := storage.NewPlanRepository(valkeyClient)
	taskRepo := storage.NewTaskRepository(valkeyClient)

	// Create a test plan
	plan, err := planRepo.Create(ctx, "test-app", "Test Plan", "Test plan description")
	req.NoError(err, "Failed to create test plan")

	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only handle POST requests to /call-tool/bulk_create_tasks
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/call-tool/bulk_create_tasks") {
			// Parse the request body
			var requestBody struct {
				PlanID string `json:"plan_id"`
				TasksJSON string `json:"tasks_json"`
			}
			err := json.NewDecoder(r.Body).Decode(&requestBody)
			if err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			// Validate required fields
			if requestBody.PlanID == "" || requestBody.TasksJSON == "" {
				http.Error(w, "Missing required fields", http.StatusBadRequest)
				return
			}

			// Parse the tasks JSON
			var taskInputs []map[string]interface{}
			err = json.Unmarshal([]byte(requestBody.TasksJSON), &taskInputs)
			if err != nil {
				http.Error(w, "Invalid tasks JSON", http.StatusBadRequest)
				return
			}

			// Convert to TaskCreateInput
			var inputs []storage.TaskCreateInput
			for _, task := range taskInputs {
				input := storage.TaskCreateInput{
					Title:       task["title"].(string),
					Description: "",
					Status:      models.TaskStatusPending,
					Priority:    models.TaskPriorityMedium,
				}

				if desc, ok := task["description"].(string); ok {
					input.Description = desc
				}

				if status, ok := task["status"].(string); ok {
					input.Status = models.TaskStatus(status)
				}

				if priority, ok := task["priority"].(string); ok {
					input.Priority = models.TaskPriority(priority)
				}

				inputs = append(inputs, input)
			}

			// Create tasks in bulk
			createdTasks, err := taskRepo.CreateBulk(ctx, requestBody.PlanID, inputs)
			if err != nil {
				http.Error(w, "Failed to create tasks: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Return the created tasks
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(createdTasks)
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create a test HTTP client
	client := server.Client()

	// Prepare the request body
	taskInputs := []map[string]interface{}{
		{
			"title":       "Task 1",
			"description": "Description for task 1",
			"priority":    "high",
		},
		{
			"title":       "Task 2",
			"description": "Description for task 2",
			"status":      "in_progress",
		},
		{
			"title": "Task 3",
		},
	}
	taskInputsJSON, err := json.Marshal(taskInputs)
	req.NoError(err, "Failed to marshal task inputs")

	requestBody := map[string]string{
		"plan_id": plan.ID,
		"tasks_json": string(taskInputsJSON),
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	req.NoError(err, "Failed to marshal request body")

	// Send the request
	resp, err := client.Post(server.URL+"/call-tool/bulk_create_tasks", "application/json", strings.NewReader(string(requestBodyJSON)))
	req.NoError(err, "Failed to send request")
	defer resp.Body.Close()

	// Check the response status code
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200")

	// Parse the response body
	var createdTasks []models.Task
	err = json.NewDecoder(resp.Body).Decode(&createdTasks)
	req.NoError(err, "Failed to parse response body")

	// Verify the created tasks
	assert := assert.New(t)
	assert.Equal(3, len(createdTasks), "Should have created 3 tasks")

	// Verify task 1
	assert.Equal("Task 1", createdTasks[0].Title)
	assert.Equal("Description for task 1", createdTasks[0].Description)
	assert.Equal(models.TaskPriorityHigh, createdTasks[0].Priority)
	assert.Equal(models.TaskStatusPending, createdTasks[0].Status) // Default status
	assert.Equal(0, createdTasks[0].Order)

	// Verify task 2
	assert.Equal("Task 2", createdTasks[1].Title)
	assert.Equal("Description for task 2", createdTasks[1].Description)
	assert.Equal(models.TaskPriorityMedium, createdTasks[1].Priority) // Default priority
	assert.Equal(models.TaskStatusInProgress, createdTasks[1].Status)
	assert.Equal(1, createdTasks[1].Order)

	// Verify task 3
	assert.Equal("Task 3", createdTasks[2].Title)
	assert.Equal("no description provided", createdTasks[2].Description) // Default description
	assert.Equal(models.TaskPriorityMedium, createdTasks[2].Priority)    // Default priority
	assert.Equal(models.TaskStatusPending, createdTasks[2].Status)       // Default status
	assert.Equal(2, createdTasks[2].Order)

	// Verify tasks are stored in Valkey
	tasks, err := taskRepo.ListByPlan(ctx, plan.ID)
	req.NoError(err, "Failed to list tasks by plan")
	assert.Equal(3, len(tasks), "Should have 3 tasks in the plan")
}

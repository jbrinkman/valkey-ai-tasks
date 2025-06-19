// Package utils provides testing utilities for the valkey-ai-tasks project
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/stretchr/testify/require"
	valkeyglide "github.com/valkey-io/valkey-glide/go/v2"
)

// TestProject creates a test project with random ID
func TestProject(name string, description string) *models.Project {
	if name == "" {
		name = fmt.Sprintf("Test Project %s", uuid.New().String()[:8])
	}

	if description == "" {
		description = fmt.Sprintf("Test project description %s", uuid.New().String()[:8])
	}

	return &models.Project{
		ID:            uuid.New().String(),
		ApplicationID: uuid.New().String(), // Add application ID for completeness
		Name:          name,
		Description:   description,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// TestTask creates a test task with random ID
func TestTask(projectID string, title string, description string) *models.Task {
	if projectID == "" {
		projectID = uuid.New().String()
	}

	if title == "" {
		title = fmt.Sprintf("Test Task %s", uuid.New().String()[:8])
	}

	if description == "" {
		description = fmt.Sprintf("Test task description %s", uuid.New().String()[:8])
	}

	return &models.Task{
		ID:          uuid.New().String(),
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		Status:      models.TaskStatusPending,
		Priority:    models.TaskPriorityMedium,
		Order:       0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// CleanupValkeyData removes all test data from Valkey
func CleanupValkeyData(ctx context.Context, t *testing.T, client *valkeyglide.Client) {
	t.Helper()
	req := require.New(t)

	// Flush all data (use with caution, only in test environments)
	_, err := client.FlushAll(ctx)
	req.NoError(err, "Failed to flush Valkey data")
}

// PopulateTestData creates test projects and tasks in Valkey
// This is useful for integration tests that need pre-populated data
func PopulateTestData(ctx context.Context, t *testing.T, client *valkeyglide.Client, projectCount int, tasksPerProject int) ([]string, []string) {
	t.Helper()
	req := require.New(t)

	projectIDs := make([]string, 0, projectCount)
	taskIDs := make([]string, 0, projectCount*tasksPerProject)

	// Create test projects
	for i := 0; i < projectCount; i++ {
		projectID := uuid.New().String()
		projectIDs = append(projectIDs, projectID)

		project := TestProject(fmt.Sprintf("Test Project %d", i), "")
		project.ID = projectID

		// Store project in Valkey (simplified, actual implementation would use repository)
		projectKey := fmt.Sprintf("project:%s", projectID)
		projectMap := project.ToMap()

		// Store project fields as hash
		result, err := client.HSet(ctx, projectKey, projectMap)
		req.NoError(err, "Failed to store project in Valkey")
		req.Greater(result, int64(0), "Expected to set at least one field")

		// Add to projects list
		result, err = client.SAdd(ctx, "projects", []string{projectID})
		req.NoError(err, "Failed to add project to projects list")

		// Create tasks for this project
		for j := 0; j < tasksPerProject; j++ {
			taskID := uuid.New().String()
			taskIDs = append(taskIDs, taskID)

			task := TestTask(projectID, fmt.Sprintf("Test Task %d-%d", i, j), "")
			task.ID = taskID
			task.Order = j

			// Store task in Valkey
			taskKey := fmt.Sprintf("task:%s", taskID)
			taskMap := task.ToMap()

			// Store task fields as hash
			result, err = client.HSet(ctx, taskKey, taskMap)
			req.NoError(err, "Failed to store task in Valkey")
			req.Greater(result, int64(0), "Expected to set at least one field")

			// Add to project tasks list
			result, err = client.SAdd(ctx, fmt.Sprintf("project:%s:tasks", projectID), []string{taskID})
			req.NoError(err, "Failed to add task to project tasks list")

			// Add to tasks list
			result, err = client.SAdd(ctx, "tasks", []string{taskID})
			req.NoError(err, "Failed to add task to tasks list")

			// Add to task order sorted set
			membersScoreMap := map[string]float64{
				taskID: float64(j),
			}
			result, err = client.ZAdd(ctx, fmt.Sprintf("project:%s:tasks:order", projectID), membersScoreMap)
			req.NoError(err, "Failed to add task to order sorted set")
		}
	}

	return projectIDs, taskIDs
}

// ToJSON converts a struct to JSON
func ToJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

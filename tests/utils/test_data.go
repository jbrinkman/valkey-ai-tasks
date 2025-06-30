// Package utils provides testing utilities for the valkey-ai-tasks project
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/internal/models"
	"github.com/stretchr/testify/require"
	valkeyglide "github.com/valkey-io/valkey-glide/go/v2"
)

// TestPlan creates a test plan with random ID
func TestPlan(name string, description string) *models.Plan {
	if name == "" {
		name = fmt.Sprintf("Test Plan %s", uuid.New().String()[:8])
	}

	if description == "" {
		description = fmt.Sprintf("Test plan description %s", uuid.New().String()[:8])
	}

	return &models.Plan{
		ID:            uuid.New().String(),
		ApplicationID: uuid.New().String(), // Add application ID for completeness
		Name:          name,
		Description:   description,
		Notes:         "",
		Status:        models.PlanStatusNew,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// TestTask creates a test task with random ID
func TestTask(planID string, title string, description string) *models.Task {
	if planID == "" {
		planID = uuid.New().String()
	}

	if title == "" {
		title = fmt.Sprintf("Test Task %s", uuid.New().String()[:8])
	}

	if description == "" {
		description = fmt.Sprintf("Test task description %s", uuid.New().String()[:8])
	}

	return &models.Task{
		ID:          uuid.New().String(),
		PlanID:      planID,
		Title:       title,
		Description: description,
		Notes:       "",
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

// PopulateTestData creates test plans and tasks in Valkey
// This is useful for integration tests that need pre-populated data
func PopulateTestData(
	ctx context.Context,
	t *testing.T,
	client *valkeyglide.Client,
	planCount int,
	tasksPerPlan int,
) ([]string, []string) {
	t.Helper()
	req := require.New(t)

	planIDs := make([]string, 0, planCount)
	taskIDs := make([]string, 0, planCount*tasksPerPlan)

	// Create test plans
	for i := 0; i < planCount; i++ {
		planID := uuid.New().String()
		planIDs = append(planIDs, planID)

		plan := TestPlan(fmt.Sprintf("Test Plan %d", i), "")
		plan.ID = planID

		// Store plan in Valkey (simplified, actual implementation would use repository)
		planKey := fmt.Sprintf("plan:%s", planID)
		planMap := plan.ToMap()

		// Store plan fields as hash
		result, err := client.HSet(ctx, planKey, planMap)
		req.NoError(err, "Failed to store plan in Valkey")
		req.Greater(result, int64(0), "Expected to set at least one field")

		// Add to plans list
		_, err = client.SAdd(ctx, "plans", []string{planID})
		req.NoError(err, "Failed to add plan to plans list")

		// Create tasks for this plan
		for j := 0; j < tasksPerPlan; j++ {
			taskID := uuid.New().String()
			taskIDs = append(taskIDs, taskID)

			task := TestTask(planID, fmt.Sprintf("Test Task %d-%d", i, j), "")
			task.ID = taskID
			task.Order = j

			// Store task in Valkey
			taskKey := fmt.Sprintf("task:%s", taskID)
			taskMap := task.ToMap()

			// Store task fields as hash
			result, err = client.HSet(ctx, taskKey, taskMap)
			req.NoError(err, "Failed to store task in Valkey")
			req.Greater(result, int64(0), "Expected to set at least one field")

			// Add to plan tasks list
			_, err = client.SAdd(ctx, fmt.Sprintf("plan:%s:tasks", planID), []string{taskID})
			req.NoError(err, "Failed to add task to plan tasks list")

			// Add to tasks list
			_, err = client.SAdd(ctx, "tasks", []string{taskID})
			req.NoError(err, "Failed to add task to tasks list")

			// Add to task order sorted set
			membersScoreMap := map[string]float64{
				taskID: float64(j),
			}
			_, err = client.ZAdd(ctx, fmt.Sprintf("plan:%s:tasks:order", planID), membersScoreMap)
			req.NoError(err, "Failed to add task to order sorted set")
		}
	}

	return planIDs, taskIDs
}

// ToJSON converts a struct to JSON
func ToJSON(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

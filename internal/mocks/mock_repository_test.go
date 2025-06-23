package mocks

import (
	"context"
	"testing"
	"time"

	"github.com/jbrinkman/valkey-ai-tasks/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockPlanRepositoryNotes(t *testing.T) {
	// Setup
	ctx := context.Background()
	repo := NewMockPlanRepository()

	// Create test plan
	testPlan := &models.Plan{
		ID:          "test-plan-id",
		Name:        "Test Plan",
		Description: "Test Description",
		Notes:       "Initial notes",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now().Add(-24 * time.Hour),
	}
	repo.plans["test-plan-id"] = testPlan

	// Test cases
	t.Run("GetNotes returns correct notes", func(t *testing.T) {
		notes, err := repo.GetNotes(ctx, "test-plan-id")
		require.NoError(t, err)
		assert.Equal(t, "Initial notes", notes)
	})

	t.Run("GetNotes returns error for non-existent plan", func(t *testing.T) {
		_, err := repo.GetNotes(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plan not found")
	})

	t.Run("UpdateNotes updates notes correctly", func(t *testing.T) {
		// Record the original updated time
		originalUpdatedAt := testPlan.UpdatedAt

		// Update notes
		err := repo.UpdateNotes(ctx, "test-plan-id", "Updated notes")
		require.NoError(t, err)

		// Verify notes were updated
		plan, err := repo.Get(ctx, "test-plan-id")
		require.NoError(t, err)
		assert.Equal(t, "Updated notes", plan.Notes)

		// Verify updated timestamp was changed
		assert.True(t, plan.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("UpdateNotes returns error for non-existent plan", func(t *testing.T) {
		err := repo.UpdateNotes(ctx, "non-existent-id", "Updated notes")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plan not found")
	})

	t.Run("UpdateNotes handles empty notes", func(t *testing.T) {
		err := repo.UpdateNotes(ctx, "test-plan-id", "")
		require.NoError(t, err)

		notes, err := repo.GetNotes(ctx, "test-plan-id")
		require.NoError(t, err)
		assert.Equal(t, "", notes)
	})

	t.Run("UpdateNotes handles special characters", func(t *testing.T) {
		specialNotes := "# Heading\n\n**Bold** and *italic*\n\n```\nCode block\n```\n\n> Quote\n\n- List item"
		err := repo.UpdateNotes(ctx, "test-plan-id", specialNotes)
		require.NoError(t, err)

		notes, err := repo.GetNotes(ctx, "test-plan-id")
		require.NoError(t, err)
		assert.Equal(t, specialNotes, notes)
	})
}

func TestMockTaskRepositoryNotes(t *testing.T) {
	// Setup
	ctx := context.Background()
	planRepo := NewMockPlanRepository()
	repo := NewMockTaskRepository(planRepo)

	// Create test task
	testTask := &models.Task{
		ID:          "test-task-id",
		PlanID:      "test-plan-id",
		Title:       "Test Task",
		Description: "Test Description",
		Notes:       "Initial task notes",
		CreatedAt:   time.Now().Add(-24 * time.Hour),
		UpdatedAt:   time.Now().Add(-24 * time.Hour),
	}
	repo.tasks["test-task-id"] = testTask

	// Test cases
	t.Run("GetNotes returns correct notes", func(t *testing.T) {
		notes, err := repo.GetNotes(ctx, "test-task-id")
		require.NoError(t, err)
		assert.Equal(t, "Initial task notes", notes)
	})

	t.Run("GetNotes returns error for non-existent task", func(t *testing.T) {
		_, err := repo.GetNotes(ctx, "non-existent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task not found")
	})

	t.Run("UpdateNotes updates notes correctly", func(t *testing.T) {
		// Record the original updated time
		originalUpdatedAt := testTask.UpdatedAt

		// Update notes
		err := repo.UpdateNotes(ctx, "test-task-id", "Updated task notes")
		require.NoError(t, err)

		// Verify notes were updated
		task, err := repo.Get(ctx, "test-task-id")
		require.NoError(t, err)
		assert.Equal(t, "Updated task notes", task.Notes)

		// Verify updated timestamp was changed
		assert.True(t, task.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("UpdateNotes returns error for non-existent task", func(t *testing.T) {
		err := repo.UpdateNotes(ctx, "non-existent-id", "Updated notes")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task not found")
	})

	t.Run("UpdateNotes handles empty notes", func(t *testing.T) {
		err := repo.UpdateNotes(ctx, "test-task-id", "")
		require.NoError(t, err)

		notes, err := repo.GetNotes(ctx, "test-task-id")
		require.NoError(t, err)
		assert.Equal(t, "", notes)
	})

	t.Run("UpdateNotes handles special characters", func(t *testing.T) {
		specialNotes := "# Task Heading\n\n**Bold** and *italic*\n\n```\nCode block\n```\n\n> Quote\n\n- List item"
		err := repo.UpdateNotes(ctx, "test-task-id", specialNotes)
		require.NoError(t, err)

		notes, err := repo.GetNotes(ctx, "test-task-id")
		require.NoError(t, err)
		assert.Equal(t, specialNotes, notes)
	})
}

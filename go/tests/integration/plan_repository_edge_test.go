package integration

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/storage"
	"github.com/jbrinkman/valkey-ai-tasks/go/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPlanRepositoryEdgeCases tests edge cases for the PlanRepository
func TestPlanRepositoryEdgeCases(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set up test context
	ctx, _, cleanup := utils.SetupValkeyTest(t)
	defer cleanup()

	// Create Valkey client
	valkeyClient, err := storage.NewValkeyClient("localhost", 6379, "", "")
	require.NoError(t, err, "Failed to create Valkey client")
	defer valkeyClient.Close()

	// Create plan repository
	planRepo := storage.NewPlanRepository(valkeyClient)

	// Test case: Get non-existent plan
	t.Run("GetNonExistentPlan", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		_, err := planRepo.Get(ctx, nonExistentID)
		assert.Error(t, err, "Expected error when getting non-existent plan")
		assert.Contains(t, err.Error(), "plan not found", "Error should indicate plan not found")
	})

	// Test case: Delete non-existent plan
	t.Run("DeleteNonExistentPlan", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		err := planRepo.Delete(ctx, nonExistentID)
		assert.Error(t, err, "Expected error when deleting non-existent plan")
		assert.Contains(t, err.Error(), "plan not found", "Error should indicate plan not found")
	})

	// Test case: Create plan with empty name
	t.Run("CreatePlanWithEmptyName", func(t *testing.T) {
		appID := "test-app-" + uuid.New().String()
		plan, err := planRepo.Create(ctx, appID, "", "Test description")
		assert.NoError(t, err, "Should be able to create plan with empty name")
		assert.Empty(t, plan.Name, "Plan name should be empty")

		// Verify plan was created
		retrievedPlan, err := planRepo.Get(ctx, plan.ID)
		assert.NoError(t, err, "Should be able to retrieve plan")
		assert.Empty(t, retrievedPlan.Name, "Retrieved plan name should be empty")
	})

	// Test case: Create plan with very long name and description
	t.Run("CreatePlanWithLongFields", func(t *testing.T) {
		appID := "test-app-" + uuid.New().String()
		longName := strings.Repeat("Very long name ", 100)        // ~1400 chars
		longDesc := strings.Repeat("Very long description ", 100) // ~2300 chars

		plan, err := planRepo.Create(ctx, appID, longName, longDesc)
		assert.NoError(t, err, "Should be able to create plan with long fields")

		// Verify plan was created with long fields
		retrievedPlan, err := planRepo.Get(ctx, plan.ID)
		assert.NoError(t, err, "Should be able to retrieve plan")
		assert.Equal(t, longName, retrievedPlan.Name, "Retrieved plan name should match")
		assert.Equal(t, longDesc, retrievedPlan.Description, "Retrieved plan description should match")
	})

	// Test case: Create plan with special characters
	t.Run("CreatePlanWithSpecialChars", func(t *testing.T) {
		appID := "test-app-" + uuid.New().String()
		specialName := "Special !@#$%^&*()_+{}|:<>?~"
		specialDesc := "Description with emoji 😀 and unicode ♠♥♦♣"

		plan, err := planRepo.Create(ctx, appID, specialName, specialDesc)
		assert.NoError(t, err, "Should be able to create plan with special characters")

		// Verify plan was created with special characters
		retrievedPlan, err := planRepo.Get(ctx, plan.ID)
		assert.NoError(t, err, "Should be able to retrieve plan")
		assert.Equal(t, specialName, retrievedPlan.Name, "Retrieved plan name should match")
		assert.Equal(t, specialDesc, retrievedPlan.Description, "Retrieved plan description should match")
	})

	// Test case: Update non-existent plan
	t.Run("UpdateNonExistentPlan", func(t *testing.T) {
		nonExistentPlan := &models.Plan{
			ID:          uuid.New().String(),
			Name:        "Non-existent Plan",
			Description: "This plan doesn't exist",
		}
		err := planRepo.Update(ctx, nonExistentPlan)
		assert.NoError(t, err, "Update should not fail for non-existent plan as it just creates it")

		// Verify the plan was actually created
		retrievedPlan, err := planRepo.Get(ctx, nonExistentPlan.ID)
		assert.NoError(t, err, "Should be able to retrieve the newly created plan")
		assert.Equal(t, nonExistentPlan.Name, retrievedPlan.Name, "Plan name should match")
	})

	// Test case: List plans with no plans
	t.Run("ListEmptyPlans", func(t *testing.T) {
		// Create a new application ID that definitely has no plans
		emptyAppID := "empty-app-" + uuid.New().String()

		plans, err := planRepo.ListByApplication(ctx, emptyAppID)
		assert.NoError(t, err, "Should not error when listing plans for empty application")
		assert.Empty(t, plans, "Should return empty slice for application with no plans")
	})

	// Test case: List plans by status
	t.Run("ListPlansByStatus", func(t *testing.T) {
		// Create a unique application ID for this test
		appID := "test-app-" + uuid.New().String()

		// Create plans with different statuses
		plan1, err := planRepo.Create(ctx, appID, "Plan New", "A new plan")
		assert.NoError(t, err, "Failed to create first test plan")
		plan1.Status = models.PlanStatusNew
		err = planRepo.Update(ctx, plan1)
		assert.NoError(t, err, "Failed to update first plan status")

		plan2, err := planRepo.Create(ctx, appID, "Plan In Progress", "A plan in progress")
		assert.NoError(t, err, "Failed to create second test plan")
		plan2.Status = models.PlanStatusInProgress
		err = planRepo.Update(ctx, plan2)
		assert.NoError(t, err, "Failed to update second plan status")

		// List plans with 'new' status
		newPlans, err := planRepo.ListByApplication(ctx, appID)
		assert.NoError(t, err, "Should not error when listing plans by application")
		assert.Equal(t, 2, len(newPlans), "Should find both plans for the application")

		// List plans with 'new' status
		newStatusPlans, err := planRepo.ListByStatus(ctx, models.PlanStatusNew)
		assert.NoError(t, err, "Should not error when listing plans by status")
		
		// Find our specific plan in the results
		foundPlan1 := false
		for _, p := range newStatusPlans {
			if p.ID == plan1.ID {
				foundPlan1 = true
				break
			}
		}
		assert.True(t, foundPlan1, "Should find plan1 in the new status plans")

		// List plans with 'in_progress' status
		inProgressPlans, err := planRepo.ListByStatus(ctx, models.PlanStatusInProgress)
		assert.NoError(t, err, "Should not error when listing plans by status")

		// Find our specific plan in the results
		foundPlan2 := false
		for _, p := range inProgressPlans {
			if p.ID == plan2.ID {
				foundPlan2 = true
				break
			}
		}
		assert.True(t, foundPlan2, "Should find plan2 in the in_progress status plans")

		// Clean up
		err = planRepo.Delete(ctx, plan1.ID)
		assert.NoError(t, err, "Failed to delete first test plan")
		err = planRepo.Delete(ctx, plan2.ID)
		assert.NoError(t, err, "Failed to delete second test plan")
	})
}

// TestPlanRepositoryConcurrentOperations tests concurrent operations on the PlanRepository
func TestPlanRepositoryConcurrentOperations(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Set up test context
	ctx, _, cleanup := utils.SetupValkeyTest(t)
	defer cleanup()

	// Create Valkey client
	valkeyClient, err := storage.NewValkeyClient("localhost", 6379, "", "")
	require.NoError(t, err, "Failed to create Valkey client")
	defer valkeyClient.Close()

	// Create plan repository
	planRepo := storage.NewPlanRepository(valkeyClient)

	// Create a test plan
	appID := "test-app-" + uuid.New().String()
	plan, err := planRepo.Create(ctx, appID, "Concurrent Test Plan", "Testing concurrent operations")
	require.NoError(t, err, "Failed to create test plan")

	// Test concurrent updates to the same plan
	t.Run("ConcurrentUpdates", func(t *testing.T) {
		// Get the plan twice (simulating two different clients)
		plan1, err := planRepo.Get(ctx, plan.ID)
		require.NoError(t, err, "Failed to get plan first time")

		plan2, err := planRepo.Get(ctx, plan.ID)
		require.NoError(t, err, "Failed to get plan second time")

		// Update the plan with the first client
		plan1.Name = "Updated by client 1"
		err = planRepo.Update(ctx, plan1)
		assert.NoError(t, err, "Failed to update plan with first client")

		// Update the plan with the second client
		plan2.Name = "Updated by client 2"
		err = planRepo.Update(ctx, plan2)
		assert.NoError(t, err, "Failed to update plan with second client")

		// Get the plan again to see the final state
		finalPlan, err := planRepo.Get(ctx, plan.ID)
		assert.NoError(t, err, "Failed to get final plan state")

		// The last update should win
		assert.Equal(t, "Updated by client 2", finalPlan.Name, "Last update should win in concurrent updates")
	})

	// Test concurrent create and delete
	t.Run("ConcurrentCreateDelete", func(t *testing.T) {
		// Create a new plan
		newAppID := "test-app-" + uuid.New().String()
		newPlan, err := planRepo.Create(ctx, newAppID, "Plan to be deleted", "This plan will be deleted")
		require.NoError(t, err, "Failed to create new plan")

		// Try to delete and update at the same time
		// First update
		newPlan.Description = "Updated description"
		err = planRepo.Update(ctx, newPlan)
		assert.NoError(t, err, "Failed to update plan")

		// Then delete
		err = planRepo.Delete(ctx, newPlan.ID)
		assert.NoError(t, err, "Failed to delete plan")

		// Try to get the deleted plan
		_, err = planRepo.Get(ctx, newPlan.ID)
		assert.Error(t, err, "Should not be able to get deleted plan")
		assert.Contains(t, err.Error(), "plan not found", "Error should indicate plan not found")
	})
}

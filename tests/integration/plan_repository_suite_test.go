package integration

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/tests/utils"
	"github.com/stretchr/testify/suite"
)

// PlanRepositorySuite is a test suite for the PlanRepository
// It includes both standard CRUD tests and edge case tests
type PlanRepositorySuite struct {
	utils.RepositoryTestSuite
}

// SetupTest sets up each test
func (s *PlanRepositorySuite) SetupTest() {
	// Call the base SetupTest to initialize the container and client
	s.RepositoryTestSuite.SetupTest()
}

// TestCreatePlan tests creating a plan
func (s *PlanRepositorySuite) TestCreatePlan() {
	planRepo := s.GetPlanRepository()

	appID := "test-app-" + uuid.New().String()
	plan, err := planRepo.Create(s.Context, appID, "Test Plan", "Test plan description")
	s.NoError(err, "Failed to create plan")
	s.NotEmpty(plan.ID, "Plan ID should not be empty")
	s.Equal("Test Plan", plan.Name, "Plan name should match")
	s.Equal("Test plan description", plan.Description, "Plan description should match")
	s.Equal(models.PlanStatusNew, plan.Status, "Plan should have default new status")
	s.Equal(appID, plan.ApplicationID, "Plan should be associated with the correct application")
}

// TestGetPlan tests retrieving a plan
func (s *PlanRepositorySuite) TestGetPlan() {
	planRepo := s.GetPlanRepository()

	// Create a plan
	appID := "test-app-" + uuid.New().String()
	plan, err := planRepo.Create(s.Context, appID, "Test Plan", "Test plan description")
	s.NoError(err, "Failed to create plan")

	// Get the plan
	retrievedPlan, err := planRepo.Get(s.Context, plan.ID)
	s.NoError(err, "Failed to get plan")
	s.Equal(plan.ID, retrievedPlan.ID, "Plan ID should match")
	s.Equal(plan.Name, retrievedPlan.Name, "Plan name should match")
	s.Equal(plan.Description, retrievedPlan.Description, "Plan description should match")
	s.Equal(plan.Status, retrievedPlan.Status, "Plan status should match")
	s.Equal(plan.ApplicationID, retrievedPlan.ApplicationID, "Plan application ID should match")
}

// TestGetNonExistentPlan tests retrieving a non-existent plan
func (s *PlanRepositorySuite) TestGetNonExistentPlan() {
	planRepo := s.GetPlanRepository()

	nonExistentID := uuid.New().String()
	_, err := planRepo.Get(s.Context, nonExistentID)
	s.Error(err, "Getting non-existent plan should return error")
	s.Contains(err.Error(), "plan not found", "Error should indicate plan not found")
}

// TestUpdatePlan tests updating a plan
func (s *PlanRepositorySuite) TestUpdatePlan() {
	planRepo := s.GetPlanRepository()

	// Create a plan
	appID := "test-app-" + uuid.New().String()
	plan, err := planRepo.Create(s.Context, appID, "Test Plan", "Test plan description")
	s.NoError(err, "Failed to create plan")

	// Update plan properties
	plan.Name = "Updated Plan Name"
	plan.Description = "Updated plan description"
	plan.Status = models.PlanStatusInProgress

	// Perform the update
	err = planRepo.Update(s.Context, plan)
	s.NoError(err, "Failed to update plan")

	// Retrieve the plan again to verify updates
	updatedPlan, err := planRepo.Get(s.Context, plan.ID)
	s.NoError(err, "Failed to get updated plan")
	s.Equal("Updated Plan Name", updatedPlan.Name, "Plan name should be updated")
	s.Equal("Updated plan description", updatedPlan.Description, "Plan description should be updated")
	s.Equal(models.PlanStatusInProgress, updatedPlan.Status, "Plan status should be updated")
}

// TestDeletePlan tests deleting a plan
func (s *PlanRepositorySuite) TestDeletePlan() {
	planRepo := s.GetPlanRepository()

	// Create a plan
	appID := "test-app-" + uuid.New().String()
	plan, err := planRepo.Create(s.Context, appID, "Test Plan", "Test plan description")
	s.NoError(err, "Failed to create plan")

	// Delete the plan
	err = planRepo.Delete(s.Context, plan.ID)
	s.NoError(err, "Failed to delete plan")

	// Try to get the deleted plan
	_, err = planRepo.Get(s.Context, plan.ID)
	s.Error(err, "Getting deleted plan should return error")
	s.Contains(err.Error(), "plan not found", "Error should indicate plan not found")
}

// TestDeleteNonExistentPlan tests deleting a non-existent plan
func (s *PlanRepositorySuite) TestDeleteNonExistentPlan() {
	planRepo := s.GetPlanRepository()

	nonExistentID := uuid.New().String()
	err := planRepo.Delete(s.Context, nonExistentID)
	s.Error(err, "Deleting non-existent plan should return error")
	s.Contains(err.Error(), "plan not found", "Error should indicate plan not found")
}

// TestListPlansByApplication tests listing plans by application
func (s *PlanRepositorySuite) TestListPlansByApplication() {
	planRepo := s.GetPlanRepository()

	// Create a unique application ID
	appID := "test-app-" + uuid.New().String()

	// Create plans
	plan1, err := planRepo.Create(s.Context, appID, "First Plan", "First plan description")
	s.NoError(err, "Failed to create first plan")

	plan2, err := planRepo.Create(s.Context, appID, "Second Plan", "Second plan description")
	s.NoError(err, "Failed to create second plan")

	// Create a plan for a different application
	otherAppID := "other-app-" + uuid.New().String()
	_, err = planRepo.Create(s.Context, otherAppID, "Other App Plan", "Plan for another application")
	s.NoError(err, "Failed to create plan for other application")

	// List plans for the first application
	plans, err := planRepo.ListByApplication(s.Context, appID)
	s.NoError(err, "Failed to list plans by application")
	s.Equal(2, len(plans), "Should have 2 plans for the application")

	// Verify plan IDs match
	planIDs := map[string]bool{
		plan1.ID: false,
		plan2.ID: false,
	}

	for _, plan := range plans {
		planIDs[plan.ID] = true
	}

	for id, found := range planIDs {
		s.True(found, "Plan with ID %s should be in the list", id)
	}
}

// TestListPlansByStatus tests listing plans by status
func (s *PlanRepositorySuite) TestListPlansByStatus() {
	planRepo := s.GetPlanRepository()

	// Create plans with different statuses
	appID := "test-app-" + uuid.New().String()

	planNew, err := planRepo.Create(s.Context, appID, "New Plan", "A new plan")
	s.NoError(err, "Failed to create new plan")
	planNew.Status = models.PlanStatusNew
	err = planRepo.Update(s.Context, planNew)
	s.NoError(err, "Failed to update new plan status")

	planInProgress, err := planRepo.Create(s.Context, appID, "In Progress Plan", "An in-progress plan")
	s.NoError(err, "Failed to create in-progress plan")
	planInProgress.Status = models.PlanStatusInProgress
	err = planRepo.Update(s.Context, planInProgress)
	s.NoError(err, "Failed to update in-progress plan status")

	// List new plans
	newPlans, err := planRepo.ListByStatus(s.Context, models.PlanStatusNew)
	s.NoError(err, "Failed to list new plans")

	// Find our specific new plan in the results
	foundNew := false
	for _, p := range newPlans {
		if p.ID == planNew.ID {
			foundNew = true
			break
		}
	}
	s.True(foundNew, "Should find the new plan in new plans list")

	// List in-progress plans
	inProgressPlans, err := planRepo.ListByStatus(s.Context, models.PlanStatusInProgress)
	s.NoError(err, "Failed to list in-progress plans")

	// Find our specific in-progress plan in the results
	foundInProgress := false
	for _, p := range inProgressPlans {
		if p.ID == planInProgress.ID {
			foundInProgress = true
			break
		}
	}
	s.True(foundInProgress, "Should find the in-progress plan in in-progress plans list")
}

// TestCreatePlanWithEmptyName tests creating a plan with an empty name
func (s *PlanRepositorySuite) TestCreatePlanWithEmptyName() {
	planRepo := s.GetPlanRepository()

	appID := "test-app-" + uuid.New().String()
	plan, err := planRepo.Create(s.Context, appID, "", "Test description")
	s.NoError(err, "Should be able to create plan with empty name")
	s.Empty(plan.Name, "Plan name should be empty")

	// Verify plan was created
	retrievedPlan, err := planRepo.Get(s.Context, plan.ID)
	s.NoError(err, "Should be able to retrieve plan")
	s.Empty(retrievedPlan.Name, "Retrieved plan name should be empty")
}

// TestCreatePlanWithEmptyDescription tests creating a plan with an empty description
func (s *PlanRepositorySuite) TestCreatePlanWithEmptyDescription() {
	planRepo := s.GetPlanRepository()

	appID := "test-app-" + uuid.New().String()
	plan, err := planRepo.Create(s.Context, appID, "Plan Title", "")
	s.NoError(err, "Should be able to create plan with empty description")
	s.Empty(plan.Description, "Plan description should be empty")
}

// TestCreatePlanWithLongFields tests creating a plan with very long name and description
func (s *PlanRepositorySuite) TestCreatePlanWithLongFields() {
	planRepo := s.GetPlanRepository()

	appID := "test-app-" + uuid.New().String()
	longName := strings.Repeat("Very long name ", 100)        // ~1400 chars
	longDesc := strings.Repeat("Very long description ", 100) // ~2300 chars

	plan, err := planRepo.Create(s.Context, appID, longName, longDesc)
	s.NoError(err, "Should be able to create plan with long fields")

	// Verify plan was created with long fields
	retrievedPlan, err := planRepo.Get(s.Context, plan.ID)
	s.NoError(err, "Should be able to retrieve plan")
	s.Equal(longName, retrievedPlan.Name, "Retrieved plan name should match")
	s.Equal(longDesc, retrievedPlan.Description, "Retrieved plan description should match")
}

// TestCreatePlanWithSpecialChars tests creating a plan with special characters
func (s *PlanRepositorySuite) TestCreatePlanWithSpecialChars() {
	planRepo := s.GetPlanRepository()

	appID := "test-app-" + uuid.New().String()
	specialName := "Special !@#$%^&*()_+{}|:<>?~"
	specialDesc := "Description with emoji 😀 and unicode ♠♥♦♣"

	plan, err := planRepo.Create(s.Context, appID, specialName, specialDesc)
	s.NoError(err, "Should be able to create plan with special characters")

	// Verify plan was created with special characters
	retrievedPlan, err := planRepo.Get(s.Context, plan.ID)
	s.NoError(err, "Should be able to retrieve plan")
	s.Equal(specialName, retrievedPlan.Name, "Retrieved plan name should match")
	s.Equal(specialDesc, retrievedPlan.Description, "Retrieved plan description should match")
}

// TestListEmptyApplication tests listing plans for an application with no plans
func (s *PlanRepositorySuite) TestListEmptyApplication() {
	planRepo := s.GetPlanRepository()

	// Create a new application ID that definitely has no plans
	emptyAppID := "empty-app-" + uuid.New().String()

	plans, err := planRepo.ListByApplication(s.Context, emptyAppID)
	s.NoError(err, "Should not error when listing plans for empty application")
	s.Empty(plans, "Should return empty slice for application with no plans")
}

// TestConcurrentUpdates tests concurrent updates to the same plan
func (s *PlanRepositorySuite) TestConcurrentUpdates() {
	planRepo := s.GetPlanRepository()

	// Create a test plan
	appID := "test-app-" + uuid.New().String()
	plan, err := planRepo.Create(s.Context, appID, "Concurrent Test Plan", "Testing concurrent operations")
	s.NoError(err, "Failed to create test plan")

	// Get the plan twice (simulating two different clients)
	plan1, err := planRepo.Get(s.Context, plan.ID)
	s.NoError(err, "Failed to get plan first time")

	plan2, err := planRepo.Get(s.Context, plan.ID)
	s.NoError(err, "Failed to get plan second time")

	// Update the plan with the first client
	plan1.Name = "Updated by client 1"
	err = planRepo.Update(s.Context, plan1)
	s.NoError(err, "Failed to update plan with first client")

	// Update the plan with the second client
	plan2.Name = "Updated by client 2"
	err = planRepo.Update(s.Context, plan2)
	s.NoError(err, "Failed to update plan with second client")

	// Get the plan again to see the final state
	finalPlan, err := planRepo.Get(s.Context, plan.ID)
	s.NoError(err, "Failed to get final plan state")

	// The last update should win
	s.Equal("Updated by client 2", finalPlan.Name, "Last update should win in concurrent updates")
}

// TestConcurrentCreateDelete tests concurrent create and delete operations
func (s *PlanRepositorySuite) TestConcurrentCreateDelete() {
	planRepo := s.GetPlanRepository()

	// Create a new plan
	newAppID := "test-app-" + uuid.New().String()
	newPlan, err := planRepo.Create(s.Context, newAppID, "Plan to be deleted", "This plan will be deleted")
	s.NoError(err, "Failed to create new plan")

	// Try to delete and update at the same time
	// First update
	newPlan.Description = "Updated description"
	err = planRepo.Update(s.Context, newPlan)
	s.NoError(err, "Failed to update plan")

	// Then delete
	err = planRepo.Delete(s.Context, newPlan.ID)
	s.NoError(err, "Failed to delete plan")

	// Try to get the deleted plan
	_, err = planRepo.Get(s.Context, newPlan.ID)
	s.Error(err, "Should not be able to get deleted plan")
	s.Contains(err.Error(), "plan not found", "Error should indicate plan not found")
}

// TestUpdateNonExistentPlan tests updating a non-existent plan
func (s *PlanRepositorySuite) TestUpdateNonExistentPlan() {
	planRepo := s.GetPlanRepository()

	nonExistentPlan := &models.Plan{
		ID:          uuid.New().String(),
		Name:        "Non-existent Plan",
		Description: "This plan doesn't exist",
	}
	err := planRepo.Update(s.Context, nonExistentPlan)
	s.NoError(err, "Update should not fail for non-existent plan as it just creates it")

	// Verify the plan was actually created
	retrievedPlan, err := planRepo.Get(s.Context, nonExistentPlan.ID)
	s.NoError(err, "Should be able to retrieve the newly created plan")
	s.Equal(nonExistentPlan.Name, retrievedPlan.Name, "Plan name should match")
}

// TestPlanRepositorySuite runs the plan repository test suite
func TestPlanRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	suite.Run(t, new(PlanRepositorySuite))
}

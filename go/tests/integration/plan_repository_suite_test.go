package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/go/tests/utils"
	"github.com/stretchr/testify/suite"
)

// PlanRepositorySuite is a test suite for the PlanRepository
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

// TestPlanRepositorySuite runs the plan repository test suite
func TestPlanRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	suite.Run(t, new(PlanRepositorySuite))
}

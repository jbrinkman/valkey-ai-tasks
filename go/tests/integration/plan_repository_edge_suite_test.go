package integration

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/go/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/go/tests/utils"
	"github.com/stretchr/testify/suite"
)

// PlanRepositoryEdgeSuite is a test suite for the PlanRepository edge cases
type PlanRepositoryEdgeSuite struct {
	utils.RepositoryTestSuite
}

// SetupTest sets up each test
func (s *PlanRepositoryEdgeSuite) SetupTest() {
	// Call the base SetupTest to initialize the container and client
	s.RepositoryTestSuite.SetupTest()
}

// TestCreatePlanWithEmptyName tests creating a plan with an empty name
func (s *PlanRepositoryEdgeSuite) TestCreatePlanWithEmptyName() {
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
func (s *PlanRepositoryEdgeSuite) TestCreatePlanWithEmptyDescription() {
	planRepo := s.GetPlanRepository()

	appID := "test-app-" + uuid.New().String()
	plan, err := planRepo.Create(s.Context, appID, "Plan Title", "")
	s.NoError(err, "Should be able to create plan with empty description")
	s.Empty(plan.Description, "Plan description should be empty")
}

// TestCreatePlanWithLongFields tests creating a plan with very long name and description
func (s *PlanRepositoryEdgeSuite) TestCreatePlanWithLongFields() {
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
func (s *PlanRepositoryEdgeSuite) TestCreatePlanWithSpecialChars() {
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
func (s *PlanRepositoryEdgeSuite) TestListEmptyApplication() {
	planRepo := s.GetPlanRepository()

	// Create a new application ID that definitely has no plans
	emptyAppID := "empty-app-" + uuid.New().String()

	plans, err := planRepo.ListByApplication(s.Context, emptyAppID)
	s.NoError(err, "Should not error when listing plans for empty application")
	s.Empty(plans, "Should return empty slice for application with no plans")
}

// TestConcurrentUpdates tests concurrent updates to the same plan
func (s *PlanRepositoryEdgeSuite) TestConcurrentUpdates() {
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
func (s *PlanRepositoryEdgeSuite) TestConcurrentCreateDelete() {
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
func (s *PlanRepositoryEdgeSuite) TestUpdateNonExistentPlan() {
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

// TestPlanRepositoryEdgeSuite runs the plan repository edge case test suite
func TestPlanRepositoryEdgeSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	suite.Run(t, new(PlanRepositoryEdgeSuite))
}

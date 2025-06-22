package mocks

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/internal/models"
	"github.com/jbrinkman/valkey-ai-tasks/internal/storage"
)

// MockPlanRepository is an in-memory implementation of PlanRepositoryInterface for testing
type MockPlanRepository struct {
	mu    sync.RWMutex
	plans map[string]*models.Plan
}

// NewMockPlanRepository creates a new mock plan repository
func NewMockPlanRepository() *MockPlanRepository {
	return &MockPlanRepository{
		plans: make(map[string]*models.Plan),
	}
}

// Create adds a new plan to the mock storage
func (r *MockPlanRepository) Create(ctx context.Context, applicationID, name, description string) (*models.Plan, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := uuid.New().String()
	plan := models.NewPlan(id, applicationID, name, description)
	r.plans[id] = plan
	return plan, nil
}

// Get retrieves a plan from the mock storage
func (r *MockPlanRepository) Get(ctx context.Context, id string) (*models.Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plan, exists := r.plans[id]
	if !exists {
		return nil, fmt.Errorf("plan not found: %s", id)
	}
	return plan, nil
}

// Update modifies an existing plan in the mock storage
func (r *MockPlanRepository) Update(ctx context.Context, plan *models.Plan) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.plans[plan.ID]
	if !exists {
		return fmt.Errorf("plan not found: %s", plan.ID)
	}

	// Update the modification time
	plan.UpdatedAt = time.Now()
	r.plans[plan.ID] = plan
	return nil
}

// Delete removes a plan from the mock storage
func (r *MockPlanRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.plans[id]
	if !exists {
		return fmt.Errorf("plan not found: %s", id)
	}
	delete(r.plans, id)
	return nil
}

// List returns all plans from the mock storage
func (r *MockPlanRepository) List(ctx context.Context) ([]*models.Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plans := make([]*models.Plan, 0, len(r.plans))
	for _, plan := range r.plans {
		plans = append(plans, plan)
	}
	return plans, nil
}

// ListByApplication returns plans for a specific application from the mock storage
func (r *MockPlanRepository) ListByApplication(ctx context.Context, applicationID string) ([]*models.Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plans := make([]*models.Plan, 0)
	for _, plan := range r.plans {
		if plan.ApplicationID == applicationID {
			plans = append(plans, plan)
		}
	}
	return plans, nil
}

// ListByStatus returns plans with a specific status from the mock storage
func (r *MockPlanRepository) ListByStatus(ctx context.Context, status models.PlanStatus) ([]*models.Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plans := make([]*models.Plan, 0)
	for _, plan := range r.plans {
		// For plans without a status field, treat them as "new" for filtering
		if plan.Status == "" {
			if status == models.PlanStatusNew {
				plans = append(plans, plan)
			}
		} else if plan.Status == status {
			plans = append(plans, plan)
		}
	}
	return plans, nil
}

// Ensure MockPlanRepository implements the PlanRepositoryInterface
var _ storage.PlanRepositoryInterface = (*MockPlanRepository)(nil)

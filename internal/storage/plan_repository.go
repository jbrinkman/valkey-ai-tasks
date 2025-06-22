package storage

import (
	"context"
	"fmt"
	"time"

	uuid "github.com/google/uuid"
	"github.com/jbrinkman/valkey-ai-tasks/internal/models"
)

// PlanRepository handles storage operations for plans
type PlanRepository struct {
	client *ValkeyClient
}

// NewPlanRepository creates a new plan repository
func NewPlanRepository(client *ValkeyClient) *PlanRepository {
	return &PlanRepository{
		client: client,
	}
}

// Create adds a new plan to the storage
func (r *PlanRepository) Create(ctx context.Context, applicationID, name, description string) (*models.Plan, error) {
	// Generate a unique ID for the plan
	id := uuid.New().String()

	// Create a new plan
	plan := models.NewPlan(id, applicationID, name, description)

	// Store the plan in Valkey
	planKey := GetPlanKey(id)
	_, err := r.client.client.HSet(ctx, planKey, plan.ToMap())
	if err != nil {
		return nil, fmt.Errorf("failed to store plan: %w", err)
	}

	// Add plan ID to the plans list
	_, err = r.client.client.SAdd(ctx, plansListKey, []string{id})
	if err != nil {
		// Try to clean up the plan if adding to the set fails
		r.client.client.Del(ctx, []string{planKey})
		return nil, fmt.Errorf("failed to add plan to list: %w", err)
	}

	// Add plan ID to the application-specific plans list
	appPlansKey := fmt.Sprintf("app:%s:plans", applicationID)
	_, err = r.client.client.SAdd(ctx, appPlansKey, []string{id})
	if err != nil {
		return nil, fmt.Errorf("failed to add plan to application list: %w", err)
	}

	return plan, nil
}

// Get retrieves a plan by ID
func (r *PlanRepository) Get(ctx context.Context, id string) (*models.Plan, error) {
	planKey := GetPlanKey(id)
	result, err := r.client.client.HGetAll(ctx, planKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve plan: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("plan not found: %s", id)
	}

	plan := &models.Plan{}
	err = plan.FromMap(result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse plan data: %w", err)
	}

	return plan, nil
}

// Update updates an existing plan
func (r *PlanRepository) Update(ctx context.Context, plan *models.Plan) error {
	// Update the updated_at timestamp
	plan.UpdatedAt = time.Now()

	// Store the updated plan in Valkey
	planKey := GetPlanKey(plan.ID)
	_, err := r.client.client.HSet(ctx, planKey, plan.ToMap())
	if err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}

	return nil
}

// Delete removes a plan and all its tasks
func (r *PlanRepository) Delete(ctx context.Context, id string) error {
	// Get the plan first to verify it exists
	plan, err := r.Get(ctx, id)
	if err != nil {
		return err
	}

	// Get all tasks for this plan
	planTasksKey := GetPlanTasksKey(id)
	taskIDs, err := r.client.client.SMembers(ctx, planTasksKey)
	if err != nil {
		return fmt.Errorf("failed to retrieve plan tasks: %w", err)
	}

	// Delete all tasks
	for taskID := range taskIDs {
		taskKey := GetTaskKey(taskID)
		_, err := r.client.client.Del(ctx, []string{taskKey})
		if err != nil {
			return fmt.Errorf("failed to delete task %s: %w", taskID, err)
		}
	}

	// Delete the plan tasks set
	_, err = r.client.client.Del(ctx, []string{planTasksKey})
	if err != nil {
		return fmt.Errorf("failed to delete plan tasks set: %w", err)
	}

	// Delete the plan
	planKey := GetPlanKey(id)
	_, err = r.client.client.Del(ctx, []string{planKey})
	if err != nil {
		return fmt.Errorf("failed to delete plan: %w", err)
	}

	// Remove the plan from the plans list
	_, err = r.client.client.SRem(ctx, plansListKey, []string{id})
	if err != nil {
		return fmt.Errorf("failed to remove plan from list: %w", err)
	}

	// Remove the plan from the application-specific plans list
	appPlansKey := fmt.Sprintf("app:%s:plans", plan.ApplicationID)
	_, err = r.client.client.SRem(ctx, appPlansKey, []string{id})
	if err != nil {
		return fmt.Errorf("failed to remove plan from application list: %w", err)
	}

	return nil
}

// List returns all plans
func (r *PlanRepository) List(ctx context.Context) ([]*models.Plan, error) {
	// Get all plan IDs
	planIDs, err := r.client.client.SMembers(ctx, plansListKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve plan IDs: %w", err)
	}

	// Get each plan
	plans := make([]*models.Plan, 0, len(planIDs))
	for id := range planIDs {
		plan, err := r.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	return plans, nil
}

// ListByStatus retrieves all plans with a specific status
func (r *PlanRepository) ListByStatus(ctx context.Context, status models.PlanStatus) ([]*models.Plan, error) {
	// Get all plan IDs
	planIDs, err := r.client.client.SMembers(ctx, plansListKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan IDs: %w", err)
	}

	var plans []*models.Plan

	// Get each plan individually
	for id := range planIDs {
		// Get the plan
		plan, err := r.Get(ctx, id)
		if err != nil {
			// Skip plans that can't be retrieved
			continue
		}

		// Check if the plan has the requested status
		if plan.Status == "" {
			// Handle plans without status (treat as "new" for filtering)
			if status != models.PlanStatusNew {
				continue
			}
		} else if plan.Status != status {
			continue
		}

		// Add plan to results
		plans = append(plans, plan)
	}

	return plans, nil
}

// ListByApplication retrieves all plans for a specific application
func (r *PlanRepository) ListByApplication(ctx context.Context, applicationID string) ([]*models.Plan, error) {
	// Get all plan IDs for this application
	appPlansKey := fmt.Sprintf("app:%s:plans", applicationID)
	planIDs, err := r.client.client.SMembers(ctx, appPlansKey)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve application plan IDs: %w", err)
	}

	// If there are no plans, return an empty slice
	if len(planIDs) == 0 {
		return []*models.Plan{}, nil
	}

	// Get each plan
	plans := make([]*models.Plan, 0, len(planIDs))
	for id := range planIDs {
		// Get the plan data
		planKey := GetPlanKey(id)
		result, err := r.client.client.HGetAll(ctx, planKey)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve plan %s: %w", id, err)
		}

		// Skip if plan doesn't exist (could have been deleted)
		if len(result) == 0 {
			continue
		}

		// Parse the plan data
		plan := &models.Plan{}
		err = plan.FromMap(result)
		if err != nil {
			return nil, fmt.Errorf("failed to parse plan data for %s: %w", id, err)
		}

		plans = append(plans, plan)
	}

	return plans, nil
}

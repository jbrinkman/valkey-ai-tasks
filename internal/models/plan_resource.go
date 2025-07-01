package models

// PlanResource represents a complete view of a plan including its tasks and notes
// This is used as a resource for the MCP server to provide a consolidated view
type PlanResource struct {
	// Plan details
	Plan *Plan `json:"plan"`

	// Tasks associated with the plan
	Tasks []*Task `json:"tasks"`
}

// NewPlanResource creates a new PlanResource with the given plan and tasks
func NewPlanResource(plan *Plan, tasks []*Task) *PlanResource {
	return &PlanResource{
		Plan:  plan,
		Tasks: tasks,
	}
}

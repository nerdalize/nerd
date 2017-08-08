package v1payload

// CreatePlanInput is the input for assigning a plan to a project.
// This results in the creation of a quota in the right namespace.
type CreatePlanInput struct {
	ProjectID   string  `json:"project_id" valid:"required"`
	PlanID      string  `json:"plan_id" valid:"required"`
	RequestsCPU float64 `json:"requests_cpu"`
}

// CreatePlanOutput is the output from assigning a plan to a project.
type CreatePlanOutput struct {
	ProjectID      string  `json:"project_id" valid:"required"`
	PlanID         string  `json:"plan_id" valid:"required"`
	RequestsCPU    float64 `json:"requests_cpu"`
	RequestsMemory float64 `json:"requests_memory"`
}

// UpdatePlanInput is the input for updating the plan capacity
type UpdatePlanInput struct {
	ProjectID      string  `json:"project_id" valid:"required"`
	PlanID         string  `json:"plan_id" valid:"required"`
	RequestsCPU    float64 `json:"requests_cpu"`
	RequestsMemory float64 `json:"requests_memory"`
}

// UpdatePlanOutput is the output for updating the plan capacity
type UpdatePlanOutput struct {
	ProjectID      string  `json:"project_id" valid:"required"`
	PlanID         string  `json:"plan_id" valid:"required"`
	RequestsCPU    float64 `json:"requests_cpu"`
	RequestsMemory float64 `json:"requests_memory"`
}

// DescribePlanInput is the input for describing a plan.
// As a plan is not necessarily always assigned to a project, the ProjectID can be empty.
type DescribePlanInput struct {
	ProjectID string `json:"project_id"`
	PlanID    string `json:"plan_id" valid:"required"`
}

// DescribePlanOutput is the output from describing a plan.
// Same as the DescribePlanInput, a plan is not always assigned to a project so the projectID can be empty.
type DescribePlanOutput struct {
	ProjectID      string  `json:"project_id"`
	PlanID         string  `json:"plan_id" valid:"required"`
	RequestsCPU    float64 `json:"requests_cpu"`
	RequestsMemory float64 `json:"requests_memory"`
	UsedCPU        float64 `json:"used_cpu"`
	UsedMemory     float64 `json:"used_memory"`
}

// RemovePlanInput is the input for removing a plan from a project
type RemovePlanInput struct {
	PlanID    string `json:"plan_id" valid:"required"`
	ProjectID string `json:"project_id" valid:"required"`
}

// RemovePlanOutput is the output from removing a plan from a project
type RemovePlanOutput struct {
}

// DeletePlanInput is the input for deleting a plan
type DeletePlanInput struct {
	PlanID string `json:"plan_id" valid:"required"`
}

// DeletePlanOutput is the output from deleting a plan
type DeletePlanOutput struct {
}

//PlanSummary is summary of a plan
type PlanSummary struct {
	ProjectID string `json:"project_id" valid:"required"`
	PlanID    string `json:"plan_id" valid:"required"`
}

// ListPlansInput is the input for listing plans.
// The ProjectID can be empty, so this would results in a list of purchased plans.
type ListPlansInput struct {
	ProjectID string `json:"project_id" valid:"required"`
}

// ListPlansOutput is the output from listing plans of a project
type ListPlansOutput struct {
	ProjectID string         `json:"project_id" valid:"required"`
	Plans     []*PlanSummary `json:"plans" valid:"required"`
}

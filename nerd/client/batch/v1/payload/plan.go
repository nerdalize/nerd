package v1payload

// CreatePlanInput is the input for assigning a plan to a project.
// This results in the creation of a quota in the right namespace.
type CreatePlanInput struct {
	PlanID       string `json:"billing_package_id"`
	ComputeUnits string `json:"compute_units" valid:"required"`
}

// CreatePlanOutput is the output from assigning a plan to a project.
type CreatePlanOutput struct {
	ProjectID    string `json:"project_id" valid:"required"`
	PlanID       string `json:"billing_package_id" valid:"required"`
	ComputeUnits string `json:"compute_units" valid:"required"`
}

// UpdatePlanInput is the input for updating the plan capacity
type UpdatePlanInput struct {
	OnDemand     bool   `json:"on_demand"`
	ComputeUnits string `json:"compute_units" valid:"required"`
}

// UpdatePlanOutput is the output for updating the plan capacity
type UpdatePlanOutput struct {
	ProjectID    string `json:"project_id" valid:"required"`
	PlanID       string `json:"billing_package_id" valid:"required"`
	ComputeUnits string `json:"compute_units" valid:"required"`
}

// RemovePlanInput is the input for removing a plan from a project
type RemovePlanInput struct {
}

// RemovePlanOutput is the output from removing a plan from a project
type RemovePlanOutput struct {
}

// DeletePlanInput is the input for deleting a plan
type DeletePlanInput struct {
}

// DeletePlanOutput is the output from deleting a plan
type DeletePlanOutput struct {
}

//PlanSummary is summary of a plan
type PlanSummary struct {
	ComputeUnits string `json:"compute_units" valid:"required"`
	PlanID       string `json:"billing_package_id" valid:"required"`
}

// ListPlansInput is the input for listing plans.
type ListPlansInput struct {
}

// ListPlansOutput is the output from listing plans of a project
type ListPlansOutput struct {
	ProjectID string         `json:"project_id" valid:"required"`
	Plans     []*PlanSummary `json:"billing_packages" valid:"required"`
	Total     *Resource
	Used      *Resource
}

// Resource is a general struct that will be used in our list payloads.
type Resource struct {
	RequestsCPU    string `json:"requests_cpu" valid:"required"`
	RequestsMemory string `json:"requests_memory" valid:"required"`
	LimitsCPU      string `json:"limits_cpu" valid:"required"`
	LimitsMemory   string `json:"limits_memory" valid:"required"`
}

package v1payload

// ListPlansOutput contains a list of plans.
type ListPlansOutput struct {
	Plans []*Plan
}

// GetPlanOutput represent a plan and its specs
type GetPlanOutput struct {
	UID     string `json:"uid"`
	URL     string `json:"url"`
	Expires string `json:"expires"`
	Usage   struct {
		Since    string `json:"since"`
		Duration string `json:"duration"`
	} `json:"usage"`
	CapacityMemory string `json:"capacity_memory"`
	CapacityCPU    string `json:"capacity_cpu"`
	Type           string `json:"type"`
}

//Plan represents a plan.
type Plan struct {
	UID string `json:"uid"`
	// Price       string `json:"price"`
	// Type        string `json:"type"`
	// ProjectSlug string `json:"project_slug"`
}

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
	Capacity struct {
		CPU    string `json:"cpu"`
		Memory string `json:"memory"`
	} `json:"capacity"`
	Type string `json:"type"`
}

//Plan represents a plan.
type Plan struct {
	UID         string `json:"uid"`
	URL         string `json:"url"`
	ProjectSlug string `json:"project_slug"`
}

package v1batch

import (
	"net/http"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

// ClientPlanInterface is an interface so client plan calls can be mocked.
type ClientPlanInterface interface {
	CreatePlan(projectID, planID string, requestsCPU float64) (output *v1payload.CreatePlanOutput, err error)
	RemovePlan(projectID, planID string) (output *v1payload.RemovePlanOutput, err error)
	DeletePlan(planID string) (output *v1payload.DeletePlanOutput, err error)
	DescribePlan(projectID, planID string) (output *v1payload.DescribePlanOutput, err error)
	ListPlans(projectID string) (output *v1payload.ListPlansOutput, err error)
	UpdatePlan(projectID, planID string, requestsCPU, requestsMemory float64) (output *v1payload.UpdatePlanOutput, err error)
}

// CreatePlan will create a plan for the precised project.
func (c *Client) CreatePlan(projectID, planID string, requestsCPU float64) (output *v1payload.CreatePlanOutput, err error) {
	output = &v1payload.CreatePlanOutput{}
	input := &v1payload.CreatePlanInput{
		ProjectID:   projectID,
		PlanID:      planID,
		RequestsCPU: requestsCPU,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, plansEndpoint), input, output)
}

// RemovePlan will delete a plan from the precised project.
func (c *Client) RemovePlan(projectID, planID string) (output *v1payload.RemovePlanOutput, err error) {
	output = &v1payload.RemovePlanOutput{}
	input := &v1payload.RemovePlanInput{
		ProjectID: projectID,
		PlanID:    planID,
	}

	return output, c.doRequest(http.MethodDelete, createPath(projectID, plansEndpoint, planID), input, output)
}

// DeletePlan will delete a plan with the provided .
func (c *Client) DeletePlan(planID string) (output *v1payload.DeletePlanOutput, err error) {
	output = &v1payload.DeletePlanOutput{}
	input := &v1payload.DeletePlanInput{
		PlanID: planID,
	}

	return output, c.doRequest(http.MethodDelete, createPath("", plansEndpoint, planID), input, output)
}

// ListPlans will return all plans for a particular project if precised.
func (c *Client) ListPlans(projectID string) (output *v1payload.ListPlansOutput, err error) {
	output = &v1payload.ListPlansOutput{}
	input := &v1payload.ListPlansInput{
		ProjectID: projectID,
	}

	return output, c.doRequest(http.MethodGet, createPath(projectID, plansEndpoint), input, output)
}

// DescribePlan returns detailed information of a plan.
func (c *Client) DescribePlan(projectID string) (output *v1payload.DescribePlanOutput, err error) {
	output = &v1payload.DescribePlanOutput{}
	input := &v1payload.DescribePlanInput{
		ProjectID: projectID,
	}

	return output, c.doRequest(http.MethodGet, createPath(projectID, plansEndpoint), input, output)
}

// UpdatePlan returns a plan with an updated cpu request.
func (c *Client) UpdatePlan(projectID, planID string, requestsCPU, requestsMemory float64) (output *v1payload.UpdatePlanOutput, err error) {
	output = &v1payload.UpdatePlanOutput{}
	input := &v1payload.UpdatePlanInput{
		ProjectID:      projectID,
		PlanID:         planID,
		RequestsCPU:    requestsCPU,
		RequestsMemory: requestsMemory,
	}

	return output, c.doRequest(http.MethodPut, createPath(projectID, plansEndpoint, planID), input, output)
}

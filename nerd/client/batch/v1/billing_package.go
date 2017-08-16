package v1batch

import (
	"net/http"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

// ClientBillingPackageInterface is an interface so client billing package calls can be mocked.
type ClientBillingPackageInterface interface {
	CreateBillingPackage(projectID, billingPackageID, requestsCPU, requestsMemory string) (output *v1payload.CreateBillingPackageOutput, err error)
	RemoveBillingPackage(projectID, billingPackageID string) (output *v1payload.RemoveBillingPackageOutput, err error)
	DeleteBillingPackage(billingPackageID string) (output *v1payload.DeleteBillingPackageOutput, err error)
	DescribeBillingPackage(projectID, billingPackageID string) (output *v1payload.DescribeBillingPackageOutput, err error)
	ListBillingPackages(projectID string) (output *v1payload.ListBillingPackagesOutput, err error)
	UpdateBillingPackage(projectID, billingPackageID, requestsCPU, requestsMemory string) (output *v1payload.UpdateBillingPackageOutput, err error)
}

// CreateBillingPackage will create a billing package for the precised project.
func (c *Client) CreateBillingPackage(projectID, billingPackageID, requestsCPU, requestsMemory string) (output *v1payload.CreateBillingPackageOutput, err error) {
	output = &v1payload.CreateBillingPackageOutput{}
	input := &v1payload.CreateBillingPackageInput{
		BillingPackageID: billingPackageID,
		RequestsCPU:      requestsCPU,
		RequestsMemory:   requestsMemory,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, billingPackagesEndpoint), input, output)
}

// RemoveBillingPackage will delete a billing package from the precised project.
func (c *Client) RemoveBillingPackage(projectID, billingPackageID string) (output *v1payload.RemoveBillingPackageOutput, err error) {
	output = &v1payload.RemoveBillingPackageOutput{}
	input := &v1payload.RemoveBillingPackageInput{}

	return output, c.doRequest(http.MethodDelete, createPath(projectID, billingPackagesEndpoint, billingPackageID), input, output)
}

// DeleteBillingPackage will delete a billing package with the provided .
func (c *Client) DeleteBillingPackage(billingPackageID string) (output *v1payload.DeleteBillingPackageOutput, err error) {
	output = &v1payload.DeleteBillingPackageOutput{}
	input := &v1payload.DeleteBillingPackageInput{}

	return output, c.doRequest(http.MethodDelete, createPath("", billingPackagesEndpoint, billingPackageID), input, output)
}

// ListBillingPackages will return all billingPackages for a particular project if precised.
func (c *Client) ListBillingPackages(projectID string) (output *v1payload.ListBillingPackagesOutput, err error) {
	output = &v1payload.ListBillingPackagesOutput{}
	input := &v1payload.ListBillingPackagesInput{}

	return output, c.doRequest(http.MethodGet, createPath(projectID, billingPackagesEndpoint), input, output)
}

// DescribeBillingPackage returns detailed information of a billing package.
func (c *Client) DescribeBillingPackage(projectID, billingPackageID string) (output *v1payload.DescribeBillingPackageOutput, err error) {
	output = &v1payload.DescribeBillingPackageOutput{}
	input := &v1payload.DescribeBillingPackageInput{}

	return output, c.doRequest(http.MethodGet, createPath(projectID, billingPackagesEndpoint, billingPackageID), input, output)
}

// UpdateBillingPackage returns a billing package with an updated cpu request.
func (c *Client) UpdateBillingPackage(projectID, billingPackageID, requestsCPU, requestsMemory string) (output *v1payload.UpdateBillingPackageOutput, err error) {
	output = &v1payload.UpdateBillingPackageOutput{}
	input := &v1payload.UpdateBillingPackageInput{
		RequestsCPU:    requestsCPU,
		RequestsMemory: requestsMemory,
	}

	return output, c.doRequest(http.MethodPut, createPath(projectID, billingPackagesEndpoint, billingPackageID), input, output)
}

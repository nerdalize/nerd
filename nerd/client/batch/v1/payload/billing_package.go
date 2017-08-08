package v1payload

// CreateBillingPackageInput is the input for assigning a billing package to a project.
// This results in the creation of a quota in the right namespace.
type CreateBillingPackageInput struct {
	ProjectID        string  `json:"project_id" valid:"required"`
	BillingPackageID string  `json:"billing_package_id" valid:"required"`
	RequestsCPU      float64 `json:"requests_cpu"`
}

// CreateBillingPackageOutput is the output from assigning a billing package to a project.
type CreateBillingPackageOutput struct {
	ProjectID        string  `json:"project_id" valid:"required"`
	BillingPackageID string  `json:"billing_package_id" valid:"required"`
	RequestsCPU      float64 `json:"requests_cpu"`
	RequestsMemory   float64 `json:"requests_memory"`
}

// UpdateBillingPackageInput is the input for updating the billing package capacity
type UpdateBillingPackageInput struct {
	ProjectID        string  `json:"project_id" valid:"required"`
	BillingPackageID string  `json:"billing_package_id" valid:"required"`
	RequestsCPU      float64 `json:"requests_cpu"`
	RequestsMemory   float64 `json:"requests_memory"`
}

// UpdateBillingPackageOutput is the output for updating the billing package capacity
type UpdateBillingPackageOutput struct {
	ProjectID        string  `json:"project_id" valid:"required"`
	BillingPackageID string  `json:"billing_package_id" valid:"required"`
	RequestsCPU      float64 `json:"requests_cpu"`
	RequestsMemory   float64 `json:"requests_memory"`
}

// DescribeBillingPackageInput is the input for describing a billing package.
// As a billingPackage is not necessarily always assigned to a project, the ProjectID can be empty.
type DescribeBillingPackageInput struct {
	ProjectID        string `json:"project_id"`
	BillingPackageID string `json:"billing_package_id" valid:"required"`
}

// DescribeBillingPackageOutput is the output from describing a billingPackage.
// Same as the DescribeBillingPackageInput, a billing package is not always
// assigned to a project so the projectID can be empty.
type DescribeBillingPackageOutput struct {
	ProjectID        string  `json:"project_id"`
	BillingPackageID string  `json:"billing_package_id" valid:"required"`
	RequestsCPU      float64 `json:"requests_cpu"`
	RequestsMemory   float64 `json:"requests_memory"`
	UsedCPU          float64 `json:"used_cpu"`
	UsedMemory       float64 `json:"used_memory"`
}

// RemoveBillingPackageInput is the input for removing a billing package from a project
type RemoveBillingPackageInput struct {
	BillingPackageID string `json:"billing_package_id" valid:"required"`
	ProjectID        string `json:"project_id" valid:"required"`
}

// RemoveBillingPackageOutput is the output from removing a billing package from a project
type RemoveBillingPackageOutput struct {
}

// DeleteBillingPackageInput is the input for deleting a billing package
type DeleteBillingPackageInput struct {
	BillingPackageID string `json:"billing_package_id" valid:"required"`
}

// DeleteBillingPackageOutput is the output from deleting a billing package
type DeleteBillingPackageOutput struct {
}

//BillingPackageSummary is summary of a billing package
type BillingPackageSummary struct {
	ProjectID        string `json:"project_id" valid:"required"`
	BillingPackageID string `json:"billing_package_id" valid:"required"`
}

// ListBillingPackagesInput is the input for listing billing packages.
// The ProjectID can be empty, so this would results in a list of purchased billing packages.
type ListBillingPackagesInput struct {
	ProjectID string `json:"project_id" valid:"required"`
}

// ListBillingPackagesOutput is the output from listing billing packages of a project
type ListBillingPackagesOutput struct {
	ProjectID       string                   `json:"project_id" valid:"required"`
	BillingPackages []*BillingPackageSummary `json:"billing_packages" valid:"required"`
}

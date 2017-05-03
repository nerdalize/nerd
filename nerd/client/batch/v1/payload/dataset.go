package v1payload

//CreateDatasetInput is used as input to dataset creation
type CreateDatasetInput struct {
	ProjectID string `json:"project_id" valid:"required"`
}

//CreateDatasetOutput is returned from creating a dataset
type CreateDatasetOutput struct {
	DatasetSummary
}

//DescribeDatasetInput is input for queue creation
type DescribeDatasetInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	DatasetID string `json:"dataset_id" valid:"required"`
}

//DescribeDatasetOutput is output for queue creation
type DescribeDatasetOutput struct {
	DatasetSummary
}

//ListDatasetsInput is input for queue creation
type ListDatasetsInput struct {
	ProjectID string `json:"project_id" valid:"required"`
}

//DatasetSummary is a small version of
type DatasetSummary struct {
	ProjectID string `json:"project_id"`
	DatasetID string `json:"dataset_id"`
	Bucket    string `json:"bucket"`
	Root      string `json:"root"`
}

//ListDatasetsOutput is output for queue creation
type ListDatasetsOutput struct {
	Datasets []*DatasetSummary
}

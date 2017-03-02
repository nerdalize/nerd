package payload

//DatasetCreateInput is used as input to dataset creation
type DatasetCreateInput struct{}

//DatasetCreateOutput is returned from creating a dataset
type DatasetCreateOutput struct {
	Dataset
	DatasetDetails
}

//DatasetDescribeOutput is returned from a specific dataset
type DatasetDescribeOutput struct {
	Dataset
	DatasetDetails
}

//DatasetListOutput is returned from the dataset listing
type DatasetListOutput struct {
	Datasets []*Dataset `json:"datasets"`
}

//DatasetDetails holds detailed information
type DatasetDetails struct {
	Root string   `json:"root"`
	Keys [][]byte `json:"keys"`
}

//Dataset is a dataset in the list output
type Dataset struct {
	ProjectID string `json:"project_id"`
	DatasetID string `json:"dataset_id"`
	Bucket    string `json:"bucket"`
}

package payload

//DatasetCreateInput is used as input to dataset creation
type DatasetCreateInput struct{}

//DatasetCreateOutput is returned from creating a dataset
type DatasetCreateOutput struct {
	Dataset
}

//DatasetDescribeOutput is returned from a specific dataset
type DatasetDescribeOutput struct {
	Dataset
}

//DatasetListOutput is returned from the dataset listing
type DatasetListOutput struct {
	Datasets []*Dataset `json:"datasets"`
}

//Dataset is a dataset in the list output
type Dataset struct {
	ProjectID string `json:"project_id"`
	DatasetID string `json:"dataset_id"`
	Bucket    string `json:"bucket"`
	Root      string `json:"root"`
}

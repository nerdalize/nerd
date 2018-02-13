//Package transfer provides primitives for uploading and downloading datasets
package transfer

// used in:
//  flex/provision_input: to download a dataset to disk (using namespace+datasetid)
//  flex/handle_output: to upload a dataset to s3 (using namespace+datasetid)

//  cmd/dataset_download: to download a dataset to disk by ID
//  cmd/dataset_upload -> uploadToDataset(): to upload a (just created) dataset to S3
//  cmd/run_job -> uploadToDataset() to create (a dataset) and upload to s3
//  cmd/opts -> to configure what implementation to use

// crd/hander -> to delete datasets and their content: Need object store directly(?)

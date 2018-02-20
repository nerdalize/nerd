package transferstore

//StoreType determines what type the object store will be
type StoreType string

const (
	//StoreTypeS3 uses a AWS S3 store
	StoreTypeS3 StoreType = "s3"
)

//StoreOptions contain options for all stores
type StoreOptions struct {
	Type StoreType `json:"type"`

	S3StoreBucket    string `json:"s3StoreBucket"`
	S3StorePrefix    string `json:"s3StorePrefix"`
	S3StoreAWSRegion string `json:"s3StoreAWSRegion"`
	S3StoreAccessKey string `json:"s3StoreAccessKey"`
	S3StoreSecretKey string `json:"s3StoreSecretKey"`
	S3SessionToken   string `json:"s3SessionToken"`
}

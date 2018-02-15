package transfer

//S3Conf configures the s3 upload/download
type S3Conf struct {
	Region       string
	AccessKey    string
	SecretKey    string
	SessionToken string
	Bucket       string
}

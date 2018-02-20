package transferarchiver

//ArchiverType determines what type the object store will be
type ArchiverType string

const (
	//ArchiverTypeTar uses the tar archiving format
	ArchiverTypeTar ArchiverType = "tar"
)

//ArchiverOptions contain options for all stores
type ArchiverOptions struct {
	Type ArchiverType `json:"type"`

	TarArchiverKeyPrefix string `json:"keyPrefix"`
}

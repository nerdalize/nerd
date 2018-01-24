package transfer

//Ref is a pointer to something stored remotely
type Ref struct {
	Location string
}

//Uploader interface can be implemented to provide upload functionality for moving
//data from a local path to a remote location
type Uploader interface {
	Upload(path string) (r *Ref, err error)
}

//Downloader provides functionality for moving data from a remote location
//to a local path
type Downloader interface {
	Download(*Ref) (path string, err error)
}

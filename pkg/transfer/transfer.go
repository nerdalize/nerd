package transfer

import "context"

//Ref is a pointer to something stored remotely
type Ref struct {
	Bucket string
	Key    string
}

//Transfer interface can be implemented to provide upload functionality for moving
//data from a local path to a remote location
//@TODO we need to be break this up into smaller interfaces for better composibility (e.g putting a dataset without reading files from the fs)
type Transfer interface {
	Upload(ctx context.Context, r *Ref, from string) (n int, err error)
	Download(ctx context.Context, r *Ref, to string) (err error)
	Delete(ctx context.Context, r *Ref) (err error)
}

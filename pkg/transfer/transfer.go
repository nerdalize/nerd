package transfer

import "context"

//Ref is a pointer to something stored remotely
type Ref struct {
	Bucket string
	Key    string
}

//Transfer interface can be implemented to provide upload functionality for moving
//data from a local path to a remote location
type Transfer interface {
	Upload(ctx context.Context, from string) (r *Ref, err error)
	Download(ctx context.Context, r *Ref, to string) (err error)
}

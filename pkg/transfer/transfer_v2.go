package transfer

import "io"

//Sync describes the process of describing
type Sync interface {
	Push(path []string) error
	Pull(path []string) error
	Close() error
}

//Trans abstracts a transfer mechanism for data
type Trans interface {
	Create(id string) (Sync, error)
	Open(id string) (Sync, error)
}

//Meta can be implemented to store transfer metadata
type Meta interface {
	Claim() error
	Release() error
}

//Objects provides an interface for object storage
type Objects interface {
	Get(key string, w io.WriterAt) error
	Put(key string, r io.Reader) error
	//@TODO add list
	//@TODO add delete
}

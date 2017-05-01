package v1payload

import "time"

//Metadata describes a dataset.
type Metadata struct {
	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`
	Size    int64     `json:"size"`
}

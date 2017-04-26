package v1payload

import "time"

//Metadata describes a dataset. It contains a header with different properties of the dataset and a
//KeyReadWriter which is used to keep track of the list of Keys (chunks) of the dataset.
type Metadata struct {
	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`
	Size    int64     `json:"size"`
}

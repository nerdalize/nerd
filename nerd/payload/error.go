package payload

//Error is used when an error needs to be returned
type Error struct {
	Message string `json:"message"`
}

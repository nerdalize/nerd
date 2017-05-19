package v1payload

//Error is the error returned by the authentication server.
type Error struct {
	Msg string `json:"error"`
}

//Error returns the error message.
func (e Error) Error() string {
	return e.Msg
}

//Cause is implemented to be compatible with the pkg/errors package.
func (e Error) Cause() error {
	return nil
}

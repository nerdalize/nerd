package v1payload

//Error is the error returned by the authentication server.
type Error struct {
	Detail string `json:"detail"`
}

//Error returns the error message.
func (e Error) Error() string {
	return e.Detail
}

//UserFacingMsg is implemeted to make sure the error is shown to the end user.
func (e Error) UserFacingMsg() string {
	return e.Error()
}

//Underlying is needed to implement the userFacing interface.
func (e Error) Underlying() error {
	return nil
}

//Cause is implemented to be compatible with the pkg/errors package.
func (e Error) Cause() error {
	return nil
}

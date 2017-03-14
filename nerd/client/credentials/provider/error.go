package provider

type UserFacingError struct {
	userFacingMsg string
	underlying    error
}

func (e UserFacingError) Error() string {
	return e.userFacingMsg
}

func (e UserFacingError) UserFacingMsg() string {
	return e.Error()
}

func (e UserFacingError) Underlying() error {
	return e.underlying
}

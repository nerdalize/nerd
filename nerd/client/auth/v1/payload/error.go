package v1payload

type Error struct {
	Detail string `json:"detail"`
}

func (e Error) Error() string {
	return e.Detail
}

func (e Error) UserFacingMsg() string {
	return e.Error()
}

func (e Error) Underlying() error {
	return nil
}

func (e Error) Cause() error {
	return nil
}

package v1payload

type AuthError struct {
	Detail string `json:"detail"`
}

func (e AuthError) Error() string {
	return e.Detail
}

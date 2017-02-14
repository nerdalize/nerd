package payload

//Error struct is returned by the API if anything goes wrong
type Error struct {
	//Retry indicates if the client can retry the request as is, this is mostly false on validation/encoding errors and true in other cases
	Retry bool `json:"retry"`

	//Message contains a overall message to the user, it should always be set to provide some feedback
	Message string `json:"message"`

	//Trace is set if the server is running in development mode, if it is empty it can be ignored
	Trace []string `json:"trace,omitempty"`

	//Fields can hold validation messages for individual fields, if empty the cause of the overal error is not due to specific field's input
	Fields map[string]string `json:"fields,omitempty"`
}

func (e Error) Error() string {
	return e.Message
}

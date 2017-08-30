package nerd

import "errors"

var (
	//ErrNotImplemented is returned when a function is not yet implemented
	ErrNotImplemented = errors.New("not yet implemented")

	//ErrTokenRevoked is returned when trying to refresh a revoked token
	ErrTokenRevoked = errors.New("ErrTokenRevoked")

	//ErrTokenUnset is returned when no oauth access token was found in the config file
	ErrTokenUnset = errors.New("You're not logged in. Please login with `nerd login`.")

	//ErrProjectIDNotSet is returned when no project id is set in the session
	ErrProjectIDNotSet = errors.New("No project ID specified, use `nerd project set` to configure a project to work on.")
)

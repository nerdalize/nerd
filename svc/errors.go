package svc

type errNoInput struct{}

func (e errNoInput) Error() string { return "no input" }

func (e errNoInput) IsNoInput() bool { return true }

//IsNoInputErr will evaluate to true if there was no input provided
func IsNoInputErr(err error) bool {
	type iface interface {
		IsNoInput() bool
	}
	te, ok := err.(iface)
	return ok && te.IsNoInput()
}

type errKubernetes struct{ error }

func (e errKubernetes) IsKubernetes() bool { return true }

//IsKubernetesErr is for unexpected kubernetes errors
func IsKubernetesErr(err error) bool {
	type iface interface {
		IsKubernetes() bool
	}
	te, ok := err.(iface)
	return ok && te.IsKubernetes()
}

type errValidation struct{ error }

func (e errValidation) IsValidation() bool { return true }

//IsValidationErr asserts for a validation error
func IsValidationErr(err error) bool {
	type iface interface {
		IsValidation() bool
	}
	te, ok := err.(iface)
	return ok && te.IsValidation()
}

type errDeadline struct{ error }

func (e errDeadline) IsDeadline() bool { return true }

//IsDeadlineErr indicates that a context deadline exceeded
func IsDeadlineErr(err error) bool {
	type iface interface {
		IsDeadline() bool
	}
	te, ok := err.(iface)
	return ok && te.IsDeadline()
}

type errAlreadyExists struct{ error }

func (e errAlreadyExists) IsAlreadyExists() bool { return true }

//IsAlreadyExistsErr indicates that what is attempted to be created already exists
func IsAlreadyExistsErr(err error) bool {
	type iface interface {
		IsAlreadyExists() bool
	}
	te, ok := err.(iface)
	return ok && te.IsAlreadyExists()
}

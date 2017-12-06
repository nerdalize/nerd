package kubevisor

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

type errNotExists struct{ error }

func (e errNotExists) IsNotExists() bool { return true }

//IsNotExistsErr indicates that what is attempted to be created already exists
func IsNotExistsErr(err error) bool {
	type iface interface {
		IsNotExists() bool
	}
	te, ok := err.(iface)
	return ok && te.IsNotExists()
}

type errNamespaceNotExists struct{ error }

func (e errNamespaceNotExists) IsNamespaceNotExists() bool { return true }

//IsNamespaceNotExistsErr indicates that what is attempted to be created already exists
func IsNamespaceNotExistsErr(err error) bool {
	type iface interface {
		IsNamespaceNotExists() bool
	}
	te, ok := err.(iface)
	return ok && te.IsNamespaceNotExists()
}

type errInvalidName struct{ error }

func (e errInvalidName) IsInvalidName() bool { return true }

//IsInvalidNameErr indicates the provided name was invalid
func IsInvalidNameErr(err error) bool {
	type iface interface {
		IsInvalidName() bool
	}
	te, ok := err.(iface)
	return ok && te.IsInvalidName()
}

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

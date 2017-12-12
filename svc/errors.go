package svc

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

type errNoLogs struct {
	reasonNoPods bool //there were no pods for the logs
}

func (e errNoLogs) IsNoLogs() bool { return true }

func (e errNoLogs) Error() string { return "no logs available" }

//IsNoLogsErr is returned when we couldn't retrieve any logs for the job
func IsNoLogsErr(err error) bool {
	type iface interface {
		IsNoLogs() bool
	}
	te, ok := err.(iface)
	return ok && te.IsNoLogs()
}

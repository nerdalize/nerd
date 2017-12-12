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

type errRaceCondition struct{ error }

func (e errRaceCondition) IsRaceCondition() bool { return true }

//IsRaceConditionErr is returned when we couldn't retrieve any logs for the job
func IsRaceConditionErr(err error) bool {
	type iface interface {
		IsRaceCondition() bool
	}
	te, ok := err.(iface)
	return ok && te.IsRaceCondition()
}

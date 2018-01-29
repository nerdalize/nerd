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

type errDatasetSpec struct{ error }

func (e errDatasetSpec) IsDatasetSpec() bool { return true }

//IsDatasetSpecErr is returned when a invalid input/output spec was given
func IsDatasetSpecErr(err error) bool {
	type iface interface {
		IsDatasetSpec() bool
	}
	te, ok := err.(iface)
	return ok && te.IsDatasetSpec()
}

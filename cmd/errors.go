package cmd

import (
	"fmt"

	"github.com/nerdalize/nerd/pkg/kubevisor"
	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
)

// errShowHelp can be returned by commands to show the commands help message next to the error.
type errShowHelp string

func (e errShowHelp) Error() string { return string(e) }

// errShowUsage can be returned by commands to show usage.
type errShowUsage string

func (e errShowUsage) Error() string { return string(e) }

func renderServiceError(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	if kubevisor.IsInvalidNameErr(err) {
		return errors.Errorf("%s: invalid name, must be an empty string or consist of alphanumeric characters, '-', '_' or '.'", fmt.Errorf(format, args...))
	}

	if kubevisor.IsDeadlineErr(err) {
		return errors.Errorf("%s: action took to long to complete, try again or check your internet connection", fmt.Errorf(format, args...))
	}

	if kubevisor.IsNotExistsErr(err) {
		return errors.Errorf("%s: it does not exist", fmt.Errorf(format, args...))
	}

	if kubevisor.IsKubernetesErr(err) {
		return errors.Errorf("%s: cluster failed to perform action: %v", fmt.Errorf(format, args...), err)
	}

	if kubevisor.IsAlreadyExistsErr(err) {
		return errors.Errorf("%s: it already exists", fmt.Errorf(format, args...))
	}

	if kubevisor.IsNamespaceNotExistsErr(err) {
		return errors.Errorf("%s: the namespace does not exist or you have no access", fmt.Errorf(format, args...))
	}

	if kubevisor.IsServiceUnavailableErr(err) {
		return errors.Errorf("%s: cluster is currently unable to receive requests, try again later", fmt.Errorf(format, args...))
	}

	if kubevisor.IsUnauthorizedErr(err) {
		return errors.Errorf("%s: you do not have permission to perform this action", fmt.Errorf(format, args...))
	}

	if svc.IsRaceConditionErr(err) {
		return errors.Errorf("%s: another process caused your action to fail, please try again", fmt.Errorf(format, args...))
	}

	return err
}

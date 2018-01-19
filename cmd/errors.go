package cmd

import (
	"fmt"

	"github.com/nerdalize/nerd/pkg/kubevisor"
	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
)

var (
	// ErrProjectNotSet is returned when no project config is found in the kube config file.
	ErrProjectNotSet = errors.New("No project set, use `nerd project set` to configure a project to work on.")
	// ErrNotLoggedIn is returned when no oauth access token was found in the config file.
	ErrNotLoggedIn = errors.New("You're not logged in. Please login with `nerd login`.")
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

	switch {
	case kubevisor.IsInvalidNameErr(err):
		return errors.Errorf("%s: invalid name, must be an empty string or consist of alphanumeric characters, '-', '_' or '.'", fmt.Errorf(format, args...))
	case kubevisor.IsDeadlineErr(err):
		return errors.Errorf("%s: action took to long to complete, try again or check your internet connection", fmt.Errorf(format, args...))
	case kubevisor.IsNetworkErr(err):
		return errors.Errorf("%s: failed to reach the cluster, make sure you're connected to the internet and try again", fmt.Errorf(format, args...))
	case kubevisor.IsNotExistsErr(err):
		return errors.Errorf("%s: it does not exist", fmt.Errorf(format, args...))
	case kubevisor.IsKubernetesErr(err):
		return errors.Errorf("%s: cluster failed to perform action: %v", fmt.Errorf(format, args...), err)
	case kubevisor.IsAlreadyExistsErr(err):
		return errors.Errorf("%s: it already exists", fmt.Errorf(format, args...))
	case kubevisor.IsNamespaceNotExistsErr(err):
		return errors.Errorf("%s: the namespace does not exist or you have no access", fmt.Errorf(format, args...))
	case kubevisor.IsServiceUnavailableErr(err):
		return errors.Errorf("%s: cluster is currently unable to receive requests, try again later", fmt.Errorf(format, args...))
	case kubevisor.IsUnauthorizedErr(err):
		return errors.Errorf("%s: you do not have permission to perform this action", fmt.Errorf(format, args...))
	case svc.IsRaceConditionErr(err):
		return errors.Errorf("%s: another process caused your action to fail, please try again", fmt.Errorf(format, args...))
	case errors.Cause(err) == ErrProjectNotSet:
		return ErrProjectNotSet
	case errors.Cause(err) == ErrNotLoggedIn:
		return ErrNotLoggedIn
	default:
		return err
	}
}

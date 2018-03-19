package cmd

import (
	"fmt"

	"github.com/nerdalize/nerd/pkg/kubevisor"
	transferstore "github.com/nerdalize/nerd/pkg/transfer/store"
	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
)

var (
	// ErrNamespaceNotSet is returned when no namespace config is found in the kube config file.
	ErrNamespaceNotSet = errors.New("no cluster set, use `nerd login` to update your configuration")
	// ErrNotLoggedIn is returned when no oauth access token was found in the config file.
	ErrNotLoggedIn = errors.New("you're not logged in. Please login with `nerd login`")
	// ErrOverwriteWarning is returned when a user is trying to use the same name for input and output datasets
	ErrOverwriteWarning = errors.New("it is not possible to use the same name for input and output datasets, as it could overwrite your dataset")
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

	err = errors.Cause(err)

	switch {
	case kubevisor.IsInvalidNameErr(err):
		return errors.Errorf("%s: invalid name, must consist of alphanumeric characters, '-' or '.'", fmt.Errorf(format, args...))
	case kubevisor.IsDeadlineErr(err):
		return errors.Errorf("%s: action took to long to complete, try again or check your internet connection", fmt.Errorf(format, args...))
	case kubevisor.IsNetworkErr(err):
		return errors.Errorf("%s: failed to reach the cluster, make sure you're connected to the internet and try again. Also, you can check the status page: http://status.nerdalize.com/", fmt.Errorf(format, args...))
	case kubevisor.IsNotExistsErr(err):
		return errors.Errorf("%s: it does not exist", fmt.Errorf(format, args...))
	case kubevisor.IsKubernetesErr(err):
		return errors.Errorf("%s: cluster failed to perform action: %v", fmt.Errorf(format, args...), err)
	case kubevisor.IsAlreadyExistsErr(err):
		return errors.Errorf("%s: this name is already in use, please use the list command to see what is already there", fmt.Errorf(format, args...))
	case kubevisor.IsNamespaceNotExistsErr(err):
		return errors.Errorf("%s: the namespace does not exist or you have no access. If the problem persists, please contact mayday@nerdalize.com.", fmt.Errorf(format, args...))
	case kubevisor.IsServiceUnavailableErr(err):
		return errors.Errorf("%s: cluster is currently unable to receive requests, try again later. Also, you can check the status page: http://status.nerdalize.com/", fmt.Errorf(format, args...))
	case kubevisor.IsUnauthorizedErr(err):
		return errors.Errorf("%s: you do not have permission to perform this action", fmt.Errorf(format, args...))
	case svc.IsRaceConditionErr(err):
		return errors.Errorf("%s: another process caused your action to fail, please try again", fmt.Errorf(format, args...))
	case errors.Cause(err) == ErrNamespaceNotSet:
		return ErrNamespaceNotSet
	case errors.Cause(err) == ErrNotLoggedIn:
		return ErrNotLoggedIn
	case errors.Cause(err) == transferstore.ErrObjectNotExists:
		return errors.Errorf("%s: dataset data is not available, it might still be uploading, check back again later", fmt.Errorf(format, args...))
	default:
		return errors.Wrapf(err, format, args...)
	}
}

func renderConfigError(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Cause(err) == ErrNotLoggedIn:
		return ErrNotLoggedIn
	case errors.Cause(err) == ErrNamespaceNotSet:
		return ErrNamespaceNotSet
	default:
		return err
	}
}

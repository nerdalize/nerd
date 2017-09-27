package command

import (
	"fmt"
	"net/url"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/command/format"
	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	"github.com/nerdalize/nerd/nerd/oauth"
	"github.com/pkg/errors"
)

var (
	planListUsage    = `nerd plan list`
	planListSynopsis = "List displays your plans."
	planListHelp     = `This command lists the plans applied to your project, but also the ones you have access to with your organization.`
)

//PlanList command
type PlanList struct {
	*command
}

//PlanListFactory returns a factory method for the join command
func PlanListFactory() (cli.Command, error) {
	comm, err := newCommand("nerd plan list", "List all your plans.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &PlanList{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *PlanList) DoRun(args []string) (err error) {
	authbase, err := url.Parse(cmd.config.Auth.APIEndpoint)
	if err != nil {
		return errors.Wrapf(err, "auth endpoint '%v' is not a valid URL", cmd.config.Auth.APIEndpoint)
	}
	authOpsClient := v1auth.NewOpsClient(v1auth.OpsClientConfig{
		Base:   authbase,
		Logger: cmd.outputter.Logger,
	})
	authClient := v1auth.NewClient(v1auth.ClientConfig{
		Base:               authbase,
		Logger:             cmd.outputter.Logger,
		OAuthTokenProvider: oauth.NewConfigProvider(authOpsClient, cmd.config.Auth.ClientID, cmd.session),
	})

	allPlans, err := authClient.ListPlans()
	if err != nil {
		return HandleError(err)
	}

	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}
	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}
	plans, err := bclient.ListPlans(ss.Project.Name)
	if err != nil {
		return HandleError(err)
	}

	for _, plan := range allPlans.Plans {
		for _, current := range plans.Plans {
			if plan.UID == current.PlanID {
				plan.UID = fmt.Sprintf("%s\t*", plan.UID)
			}
		}
	}
	header := "PLANS"
	pretty := "{{range $i, $x := $.Plans}}{{$x.UID}}\n{{end}}"
	raw := "{{range $i, $x := $.Plans}}{{$x.UID}}\t{{$x.CPU}}\n{{end}}"
	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(plans, header, pretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(plans, raw),
		format.OutputTypeJSON:   format.NewJSONDecorator(plans.Plans),
	})

	return nil
}

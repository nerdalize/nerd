package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/pkg/transfer"
	"github.com/nerdalize/nerd/svc"
)

//JobRun command
type JobRun struct {
	KubeOpts
	TransferOpts
	Name    string   `long:"name" short:"n" description:"assign a name to the job"`
	Env     []string `long:"env" short:"e" description:"environment variables to use"`
	Inputs  []string `long:"input" description:"specify one or more inputs that will be downloaded for the job"`
	Outputs []string `long:"output" description:"specify one or more output folders that will be uploaded as datasets after the job is finished"`

	*command
}

//JobRunFactory creates the command
func JobRunFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &JobRun{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, flags.PassAfterNonOption)
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *JobRun) Execute(args []string) (err error) {
	if len(args) < 1 {
		return errShowUsage(MessageNotEnoughArguments)
	}

	kopts := cmd.KubeOpts
	deps, err := NewDeps(cmd.Logger(), kopts)
	if err != nil {
		return renderConfigError(err, "failed to configure")
	}

	//setup a context with a timeout
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

	//setup job arguments
	jargs := []string{}
	if len(args) > 1 {
		jargs = args[1:]
	}

	jenv := map[string]string{}
	for _, l := range cmd.Env {
		split := strings.SplitN(l, "=", 2)
		if len(split) < 2 {
			return fmt.Errorf("invalid environment variable format, expected 'FOO=bar' format, got: %v", l)
		}
		jenv[split[0]] = split[1]
	}

	//setup the transfer manager
	kube := svc.NewKube(deps)
	mgr, sto, sta, err := cmd.TransferOpts.TransferManager(kube)
	if err != nil {
		return errors.Wrap(err, "failed to setup transfer manager")
	}

	//keep handles to update the job froms and to
	inputs := []transfer.Handle{}
	outputs := []transfer.Handle{}

	//start with input volumes
	vols := map[string]*svc.JobVolume{}
	for _, input := range cmd.Inputs {
		parts := strings.Split(input, ":")
		if len(parts) != 2 {
			return fmt.Errorf("invalid input specified, expected '<DIR|DATASET_ID>:<JOB_DIR>' format, got: %s", input)
		}

		if !filepath.IsAbs(parts[1]) {
			return fmt.Errorf("the job directory for the input dataset must be provided as an absolute path")
		}

		//detect if the users tries to refer to a path to upload but it is not absolute
		if strings.Contains(parts[0], string(filepath.Separator)) && !filepath.IsAbs(parts[0]) {
			return fmt.Errorf("when providing a local directory as input it must be given as an absolute path")
		}

		//if the first part can be considered a path, upload it immediately
		var h transfer.Handle
		if filepath.IsAbs(parts[0]) { //create (and upload) a new dataset
			h, err = mgr.Create(ctx, "", *sto, *sta)
			if err != nil {
				return errors.Wrap(err, "failed to create dataset")
			}

			//@TODO extend ctx deadline

			err = h.Push(ctx, parts[0], transfer.NewDiscardReporter())
			if err != nil {
				return errors.Wrap(err, "failed to update dataset")
			}

			cmd.out.Infof("Uploaded input dataset: '%s'", h.Name())
		} else { //open an existing dataset
			h, err = mgr.Open(ctx, parts[0])
			if err != nil {
				return errors.Wrap(err, "failed to open dataset")
			}

		}

		//add handler for job mapping
		inputs = append(inputs, h)
		defer h.Close()

		vols[parts[1]] = &svc.JobVolume{
			MountPath:    parts[1],
			InputDataset: h.Name(),
		}
	}

	// var outputDataset string
	for _, output := range cmd.Outputs {
		parts := strings.Split(output, ":")
		if len(parts) < 1 || len(parts) > 2 {
			return fmt.Errorf("invalid output specified, expected '<JOB_DIR>:[DATASET_NAME]' format, got: %s", output)
		}

		if !filepath.IsAbs(parts[0]) {
			return fmt.Errorf("the job directory for the output dataset must be provided as an absolute path")
		}

		vol, ok := vols[parts[0]]
		if !ok {
			vol = &svc.JobVolume{MountPath: parts[0]}
			vols[parts[0]] = vol
		}

		//if the second part is provided we want to upload the output to a specific  dataset
		var h transfer.Handle
		if len(parts) == 2 { //open an existing dataset
			h, err = mgr.Open(ctx, parts[1])
			if err != nil {
				return errors.Wrap(err, "failed to open dataset")
			}

		} else { //create an empty dataset for the output
			h, err = mgr.Create(ctx, "", *sto, *sta)
			if err != nil {
				return errors.Wrap(err, "failed to create dataset")
			}

			cmd.out.Infof("Setup empty output dataset: '%s'", h.Name())
		}

		//register for job mapping and cleanup
		outputs = append(outputs, h)
		defer h.Close()

		vol.OutputDataset = h.Name()
	}

	//continue with actuall creating the job
	in := &svc.RunJobInput{
		Image: args[0],
		Name:  cmd.Name,
		Env:   jenv,
		Args:  jargs,
	}

	for _, vol := range vols {
		in.Volumes = append(in.Volumes, *vol)
	}

	out, err := kube.RunJob(ctx, in)
	if err != nil {
		return renderServiceError(err, "failed to run job")
	}

	//add job to each dataset's InputFor
	for _, h := range inputs {
		_, err := kube.UpdateDataset(ctx, &svc.UpdateDatasetInput{Name: h.Name(), InputFor: out.Name})
		if err != nil {
			return err
		}
	}

	//add job to each dataset's OutputOf
	for _, h := range outputs {
		_, err := kube.UpdateDataset(ctx, &svc.UpdateDatasetInput{Name: h.Name(), OutputFrom: out.Name})
		if err != nil {
			return err
		}
	}

	cmd.out.Infof("Submitted job: '%s'", out.Name)
	cmd.out.Infof("To see whats happening, use: 'nerd job list'")
	return nil
}

// Description returns long-form help text
func (cmd *JobRun) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *JobRun) Synopsis() string { return "Runs a job on your compute cluster" }

// Usage shows usage
func (cmd *JobRun) Usage() string { return "nerd job run [OPTIONS] IMAGE [COMMAND] [ARG...]" }

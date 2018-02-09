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

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

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

	kube := svc.NewKube(deps)

	//start with input volumes
	//@TODO move this logic to a separate package and test is
	var inputDataset string
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

		//if the input spec has an absolute path as its first part, try to upload it for the user
		var bucket string
		var key string
		if filepath.IsAbs(parts[0]) {

			//the user has provided a path as its input
			var trans transfer.Transfer
			trans, err = cmd.TransferOpts.Transfer()
			if err != nil {
				return errors.Wrap(err, "failed configure transfer")
			}

			var ref *transfer.Ref
			ref, inputDataset, err = uploadToDataset(ctx, trans, cmd.AWSS3Bucket, kube, parts[0], "")
			if err != nil {
				return err
			}

			cmd.out.Infof("Uploaded input dataset: '%s'", inputDataset)
			bucket = ref.Bucket
			key = ref.Key
		} else {

			//the user (probably) has provided a dataset id as input
			var out *svc.GetDatasetOutput
			out, err = kube.GetDataset(ctx, &svc.GetDatasetInput{Name: parts[0]})
			if err != nil {
				return errors.Wrapf(err, "failed to get dataset '%s' ", parts[0])
			}

			if out.Bucket == "" || out.Key == "" {
				return errors.Errorf("the dataset '%s' cannot be used as input it has no key and/or bucket configured", parts[0])
			}
			bucket = out.Bucket
			key = out.Key
			inputDataset = out.Name
		}

		vols[parts[1]] = &svc.JobVolume{
			MountPath: parts[1],
			Input: &transfer.Ref{
				Bucket: bucket,
				Key:    key,
			},
		}
	}

	var outputDataset string
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

		if len(parts) == 2 {
			var out *svc.GetDatasetOutput
			out, err = kube.GetDataset(ctx, &svc.GetDatasetInput{Name: parts[1]})
			if err != nil {
				return errors.Wrapf(err, "failed to get dataset '%s' ", parts[1]) //@TODO do we want to hint the user that he might meant to specify a releative directory(?)
			}

			if out.Bucket == "" || out.Key == "" {
				return errors.Errorf("the dataset '%s' cannot be used as output as it has no key and/or bucket configured", parts[0])
			}

			vol.Output = &transfer.Ref{
				Key:    out.Key,
				Bucket: out.Bucket,
			}
			outputDataset = out.Name
		} else {
			var trans transfer.Transfer
			trans, err = cmd.TransferOpts.Transfer()
			if err != nil {
				return errors.Wrap(err, "failed configure transfer")
			}

			var ref *transfer.Ref
			ref, outputDataset, err = uploadToDataset(ctx, trans, cmd.AWSS3Bucket, kube, "", "")
			if err != nil {
				return err
			}

			cmd.out.Infof("Setup empty output dataset: '%s'", outputDataset)
			vol.Output = &transfer.Ref{
				Key:    ref.Key,
				Bucket: ref.Bucket,
			}
		}
	}

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

	err = updateDataset(ctx, inputDataset, outputDataset, out.Name, kube)
	cmd.out.Infof("Submitted job: '%s'", out.Name)
	cmd.out.Infof("To see whats happening, use: 'nerd job list'")
	return nil
}

func updateDataset(ctx context.Context, inputDataset, outputDataset, job string, kube *svc.Kube) error {
	_, err := kube.UpdateDataset(ctx, &svc.UpdateDatasetInput{Name: inputDataset, InputFor: job})
	if err != nil {
		return err
	}
	_, err = kube.UpdateDataset(ctx, &svc.UpdateDatasetInput{Name: outputDataset, OutputFrom: job})
	if err != nil {
		return err
	}
	return nil
}

// Description returns long-form help text
func (cmd *JobRun) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *JobRun) Synopsis() string { return "Runs a job on your compute cluster" }

// Usage shows usage
func (cmd *JobRun) Usage() string { return "nerd job run [OPTIONS] IMAGE [COMMAND] [ARG...]" }

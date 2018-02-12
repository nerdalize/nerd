package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

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
	var inputDatasets []string
	vols := map[string]*svc.JobVolume{}
	for _, input := range cmd.Inputs {
		var inputDataset string

		parts := strings.Split(input, ":")

		//Two accepted cases:
		//- Two unix paths with a colon separating them, e.g. ~/data:/input
		//- Windows path with a disk specification, e.g. C:/data:/input
		if len(parts) != 2 && len(parts) != 3 {
			return fmt.Errorf("invalid input specified, expected '<DIR|DATASET_ID>:<JOB_DIR>' format, got: %s", input)
		}

		//Handle Windows paths where DIR may contain colons
		//e.g. C:/foo/bar:/input will be parsed into []string{"C", "/foo/bar", "/input"}
		//and should be turned into []string{"C:/foo/bar", "/input"}
		//We assume that POSIX paths will never have colons
		parts = []string{strings.Join(parts[:len(parts)-1], ":"), parts[len(parts)-1]}

		//Expand tilde for homedir
		parts[0], err = homedir.Expand(parts[0])
		if err != nil {
			return errors.Wrap(err, "failed to expand home directory in dataset local path")
		}

		//Normalize all slashes to native platform slashes (e.g. / to \ on Windows)
		parts[0] = filepath.FromSlash(parts[0])

		//if the input spec has a path-like string, try to upload it for the user
		// var bucket string
		// var key string
		if strings.Contains(parts[0], string(filepath.Separator)) {
			//the user has provided a path as its input, clean it and make it absolute
			parts[0], err = filepath.Abs(parts[0])
			if err != nil {
				return errors.Wrap(err, "failed to turn local dataset path into absolute path")
			}

			var trans transfer.Transfer
			trans, err = cmd.TransferOpts.Transfer()
			if err != nil {
				return errors.Wrap(err, "failed configure transfer")
			}

			// var ref *transfer.Ref
			_, inputDataset, err = uploadToDataset(ctx, trans, cmd.AWSS3Bucket, kube, parts[0], "")
			if err != nil {
				return err
			}

			cmd.out.Infof("Uploaded input dataset: '%s'", inputDataset)
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

			inputDataset = out.Name
		}

		inputDatasets = append(inputDatasets, inputDataset)

		vols[parts[1]] = &svc.JobVolume{
			MountPath:    parts[1],
			InputDataset: inputDataset,
		}

		err = deps.val.Struct(vols[parts[1]])
		if err != nil {
			return errors.Wrap(err, "incorrect input")
		}
	}

	var outputDatasets []string
	for _, output := range cmd.Outputs {
		var outputDataset string

		parts := strings.Split(output, ":")
		if len(parts) < 1 || len(parts) > 2 {
			return fmt.Errorf("invalid output specified, expected '<JOB_DIR>:[DATASET_NAME]' format, got: %s", output)
		}

		vol, ok := vols[parts[0]]
		if !ok {
			vol = &svc.JobVolume{MountPath: parts[0]}
			vols[parts[0]] = vol
		}

		err = deps.val.Struct(vol)
		if err != nil {
			return errors.Wrap(err, "incorrect output")
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

			outputDataset = out.Name
		} else {
			var trans transfer.Transfer
			trans, err = cmd.TransferOpts.Transfer()
			if err != nil {
				return errors.Wrap(err, "failed configure transfer")
			}

			_, outputDataset, err = uploadToDataset(ctx, trans, cmd.AWSS3Bucket, kube, "", "")
			if err != nil {
				return err
			}

			cmd.out.Infof("Setup empty output dataset: '%s'", outputDataset)
		}

		outputDatasets = append(outputDatasets, outputDataset)

		vol.OutputDataset = outputDataset
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

	//Register datasets as being used
	for _, inputDataset := range inputDatasets {
		_, err := kube.UpdateDataset(ctx, &svc.UpdateDatasetInput{Name: inputDataset, InputFor: out.Name})
		if err != nil {
			return errors.Wrapf(err, "failed to update input dataset '%s'", inputDataset)
		}
	}
	for _, outputDataset := range outputDatasets {
		_, err := kube.UpdateDataset(ctx, &svc.UpdateDatasetInput{Name: outputDataset, OutputFrom: out.Name})
		if err != nil {
			return errors.Wrapf(err, "failed to update output dataset '%s'", outputDataset)
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

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	Name       string   `long:"name" short:"n" description:"assign a name to the job"`
	Env        []string `long:"env" short:"e" description:"environment variables to use"`
	Memory     string   `long:"memory" short:"m" description:"memory to use for this job, expressed in gigabytes" default:"3"`
	VCPU       string   `long:"vcpu" description:"number of vcpus to use for this job" default:"2"`
	Inputs     []string `long:"input" description:"specify one or more inputs that will be used for the job using the following format: <DIR|DATASET_NAME>:<JOB_DIR>"`
	Outputs    []string `long:"output" description:"specify one or more output folders that will be stored as datasets after the job is finished using the following format: <DATASET_NAME>:<JOB_DIR>"`
	Private    bool     `long:"private" description:"use this flag with a private image, a prompt will ask for your username and password of the repository that stores the image. If NERD_IMAGE_USERNAME and/or NERD_IMAGE_PASSWORD environment variables are set, those values are used instead."`
	CleanCreds bool     `long:"clean-creds" description:"to be used with the '--private' flag, a prompt will ask again for your image repository username and password. If NERD_IMAGE_USERNAME and/or NERD_IMAGE_PASSWORD environment variables are provided, they will be used as values to update the secret."`
	*command
}

//JobRunFactory creates the command
func JobRunFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &JobRun{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, &TransferOpts{}, flags.PassAfterNonOption, "nerd job run")
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//ParseInputSpecification will look at an input string and return its parts if valid
func ParseInputSpecification(input string) (parts []string, err error) {
	parts = strings.Split(input, ":")

	//Two accepted cases:
	//- Two unix paths with a colon separating them, e.g. ~/data:/input
	//- Windows path with a disk specification, e.g. C:/data:/input
	if len(parts) != 2 && len(parts) != 3 {
		return nil, fmt.Errorf("invalid input specified, expected '<DIR|DATASET_ID>:<JOB_DIR>' format, got: %s", input)
	}

	//Handle Windows paths where DIR may contain colons
	//e.g. C:/foo/bar:/input will be parsed into []string{"C", "/foo/bar", "/input"}
	//and should be turned into []string{"C:/foo/bar", "/input"}
	//We assume that POSIX paths will never have colons
	parts = []string{strings.Join(parts[:len(parts)-1], ":"), parts[len(parts)-1]}

	//Expand tilde for homedir
	parts[0], err = homedir.Expand(parts[0])
	if err != nil {
		return nil, errors.Wrap(err, "failed to expand home directory in dataset local path")
	}

	//Normalize all slashes to native platform slashes (e.g. / to \ on Windows)
	parts[0] = filepath.FromSlash(parts[0])

	// Ensure that all parts are non-empty
	if len(strings.TrimSpace(parts[0])) == 0 {
		return nil, errors.New("input source is empty")
	} else if len(strings.TrimSpace(parts[1])) == 0 {
		return nil, errors.New("input mount path is empty")
	}

	return parts, nil
}

// dsHandle keeps handles to update the job froms and to.
// newDs helps us to keep track if the dataset used is an ad-hoc dataset or a dataset that was previously submitted,
// so we know which ones we should delete in case of problems with `nerd job run`.
type dsHandle struct {
	handle transfer.Handle
	newDs  bool
}

//Execute runs the command
func (cmd *JobRun) Execute(args []string) (err error) {
	if len(args) < 1 {
		return errShowUsage(fmt.Sprintf(MessageNotEnoughArguments, 1, ""))
	}

	kopts := cmd.globalOpts.KubeOpts
	deps, err := NewDeps(cmd.Logger(), kopts)
	if err != nil {
		return renderConfigError(err, "failed to configure")
	}

	//setup a context with a timeout
	ctx := context.TODO()

	err = checkResources(cmd.Memory, cmd.VCPU)
	if err != nil {
		return err
	}
	err = compareNames(cmd.Inputs, cmd.Outputs)
	if err != nil {
		return err
	}

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
	t, ok := cmd.advancedOpts.(*TransferOpts)
	if !ok {
		return fmt.Errorf("unable to use transfer options")
	}
	mgr, sto, sta, err := t.TransferManager(kube)
	if err != nil {
		return errors.Wrap(err, "failed to setup transfer manager")
	}

	//keep handles to update the job froms and to
	inputs := []dsHandle{}
	outputs := []dsHandle{}

	//start with input volumes
	vols := map[string]*svc.JobVolume{}
	for _, input := range cmd.Inputs {
		var parts []string
		parts, err = ParseInputSpecification(input)
		if err != nil {
			return cmd.rollbackDatasets(ctx, mgr, inputs, outputs, errors.Wrap(err, "failed to parse input specification"))
		}

		//if the input spec has a path-like string, try to upload it for the user
		var h dsHandle
		if strings.Contains(parts[0], string(filepath.Separator)) {
			//the user has provided a path as its input, clean it and make it absolute
			parts[0], err = filepath.Abs(parts[0])
			if err != nil {
				return cmd.rollbackDatasets(ctx, mgr, inputs, outputs, errors.Wrap(err, "failed to turn local dataset path into absolute path"))
			}

			h.handle, err = mgr.Create(ctx, "", *sto, *sta)
			if err != nil {
				return renderServiceError(
					cmd.rollbackDatasets(ctx, mgr, inputs, outputs, err),
					"failed to create dataset",
				)
			}

			h.newDs = true
			err = h.handle.Push(ctx, parts[0], &progressBarReporter{})
			if err != nil {
				return renderServiceError(
					cmd.rollbackDatasets(ctx, mgr, append(inputs, h), outputs, err),
					"failed to upload dataset",
				)
			}

			cmd.out.Infof("Uploaded input dataset: '%s'", h.handle.Name())
		} else { //open an existing dataset
			h.handle, err = mgr.Open(ctx, parts[0])
			if err != nil {
				return renderServiceError(
					cmd.rollbackDatasets(ctx, mgr, inputs, outputs, err),
					"failed to open dataset '%s'", parts[0],
				)
			}
			h.newDs = false
		}

		//add handler for job mapping
		inputs = append(inputs, h)
		defer h.handle.Close()

		vols[parts[1]] = &svc.JobVolume{
			MountPath:    parts[1],
			InputDataset: h.handle.Name(),
		}

		err = deps.val.Struct(vols[parts[1]])
		if err != nil {
			return cmd.rollbackDatasets(ctx, mgr, inputs, outputs, errors.Wrap(err, "incorrect input"))
		}
	}

	for _, output := range cmd.Outputs {
		parts := strings.Split(output, ":")
		if len(parts) < 1 || len(parts) > 2 {
			return cmd.rollbackDatasets(ctx, mgr, inputs, outputs, fmt.Errorf("invalid output specified, expected '<JOB_DIR>:[DATASET_NAME]' format, got: %s", output))
		}

		vol, ok := vols[parts[len(parts)-1]]
		if !ok {
			vol = &svc.JobVolume{MountPath: parts[len(parts)-1]}
			vols[parts[len(parts)-1]] = vol
		}

		err = deps.val.Struct(vol)
		if err != nil {
			return cmd.rollbackDatasets(ctx, mgr, inputs, outputs, errors.Wrap(err, "incorrect output"))
		}

		//if the second part is provided we want to upload the output to a specific  dataset
		var h dsHandle
		if len(parts) == 2 { //open an existing dataset
			h.handle, err = mgr.Open(ctx, parts[0])
			h.newDs = false
			// @TODO check if the error is "dataset doesn't exist", and only in this case we should create a new one
			if err != nil {
				h.handle, err = mgr.Create(ctx, parts[0], *sto, *sta)
				if err != nil {
					return renderServiceError(
						cmd.rollbackDatasets(ctx, mgr, inputs, outputs, err),
						"failed to open/create dataset '%s'", parts[0],
					)
				}
				cmd.out.Infof("Setup empty output dataset: '%s'", h.handle.Name())
			}

		} else { //create an empty dataset for the output
			h.newDs = true
			h.handle, err = mgr.Create(ctx, "", *sto, *sta)
			if err != nil {
				return renderServiceError(
					cmd.rollbackDatasets(ctx, mgr, inputs, append(outputs, h), err),
					"failed to create dataset",
				)
			}

			cmd.out.Infof("Setup empty output dataset: '%s'", h.handle.Name())
		}

		//register for job mapping and cleanup
		outputs = append(outputs, h)
		defer h.handle.Close()

		vol.OutputDataset = h.handle.Name()
	}

	//continue with actuall creating the job
	in := &svc.RunJobInput{
		Image:  args[0],
		Name:   cmd.Name,
		Env:    jenv,
		Args:   jargs,
		Memory: fmt.Sprintf("%sGi", cmd.Memory),
		VCPU:   cmd.VCPU,
	}
	if cmd.Private {
		secrets, err := kube.ListSecrets(ctx, &svc.ListSecretsInput{})
		if err != nil {
			return renderServiceError(err, "failed to list secrets")
		}
		_, _, registry := svc.ExtractRegistry(in.Image)
		for _, secret := range secrets.Items {
			if secret.Details.Image == in.Image {
				if cmd.CleanCreds {
					username, password, err := cmd.getCredentials(registry)
					if err != nil {
						return err
					}
					_, err = kube.UpdateSecret(ctx, &svc.UpdateSecretInput{Name: secret.Name, Username: username, Password: password})
					if err != nil {
						return renderServiceError(err, "failed to update secret")
					}
				}
				in.Secret = secret.Name
				break
			}
		}
		if in.Secret == "" {
			username, password, err := cmd.getCredentials(registry)
			if err != nil {
				return err
			}
			secret, err := kube.CreateSecret(ctx, &svc.CreateSecretInput{
				Image:    in.Image,
				Username: username,
				Password: password,
			})
			if err != nil {
				return renderServiceError(err, "failed to create secret")
			}
			in.Secret = secret.Name
		}
	}

	for _, vol := range vols {
		in.Volumes = append(in.Volumes, *vol)
	}

	out, err := kube.RunJob(ctx, in)
	if err != nil {
		cmd.rollbackDatasets(ctx, mgr, inputs, outputs, nil)
		return renderServiceError(err, "failed to run job")
	}

	err = updateDatasets(ctx, kube, inputs, outputs, out.Name)
	if err != nil {
		return err
	}

	cmd.out.Infof("Submitted job: '%s'", out.Name)
	cmd.out.Infof("To see whats happening, use: 'nerd job list'")
	return nil
}

func checkResources(memory, vcpu string) error {
	if memory != "" {
		m, err := strconv.ParseFloat(memory, 64)
		if err != nil {
			return fmt.Errorf("invalid memory option format, %v", err)
		}
		if m > 60 || m <= 0 {
			return fmt.Errorf("invalid value for memory parameter. Memory request must be greater than 0 and lower than 60Gbs")
		}
	}
	if vcpu != "" {
		v, err := strconv.ParseFloat(vcpu, 64)
		if err != nil {
			return fmt.Errorf("invalid vcpu option format, %v", err)
		}
		if v > 40 || v <= 0 {
			return fmt.Errorf("invalid value for vcpu parameter. VCPU request must be greater than 0 and lower than 40")
		}
	}
	return nil
}

func (cmd *JobRun) rollbackDatasets(ctx context.Context, mgr transfer.Manager, inputs, outputs []dsHandle, err error) error {
	for _, input := range inputs {
		if input.newDs {
			mgr.Remove(ctx, input.handle.Name())
		}
	}
	for _, output := range outputs {
		if output.newDs {
			mgr.Remove(ctx, output.handle.Name())
		}
	}

	return err
}

func (cmd *JobRun) getCredentials(registry string) (username, password string, err error) {
	if registry == "index.docker.io" {
		registry = "Docker Hub"
	}
	cmd.out.Infof("Please provide credentials for the %s repository that stores the private image:", registry)
	username = os.Getenv("NERD_IMAGE_USERNAME")
	if username == "" {
		username, err = cmd.out.Ask("Username: ")
		if err != nil {
			return username, password, err
		}
	}
	password = os.Getenv("NERD_IMAGE_PASSWORD")
	if password == "" {
		password, err = cmd.out.AskSecret("Password: ")
		if err != nil {
			return username, password, err
		}
	}
	return username, password, err
}

func updateDatasets(ctx context.Context, kube *svc.Kube, inputs, outputs []dsHandle, name string) error {
	//add job to each dataset's InputFor
	for _, input := range inputs {
		_, err := kube.UpdateDataset(ctx, &svc.UpdateDatasetInput{Name: input.handle.Name(), InputFor: name})
		if err != nil {
			return err
		}
	}
	//add job to each dataset's OutputOf
	for _, output := range outputs {
		_, err := kube.UpdateDataset(ctx, &svc.UpdateDatasetInput{Name: output.handle.Name(), OutputFrom: name})
		if err != nil {
			return err
		}
	}
	return nil
}

func compareNames(inputs, outputs []string) error {
	for i := range inputs {
		input := strings.Split(inputs[i], ":")
		for o := range outputs {
			output := strings.Split(outputs[o], ":")
			if input[0] == output[0] {
				return ErrOverwriteWarning
			}
		}
	}
	return nil
}

// Description returns long-form help text
func (cmd *JobRun) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *JobRun) Synopsis() string { return "Runs a job on your compute cluster" }

// Usage shows usage
func (cmd *JobRun) Usage() string { return "nerd job run [OPTIONS] IMAGE [ARG...]" }

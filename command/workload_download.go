package command

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/aws"
	v1datatransfer "github.com/nerdalize/nerd/nerd/service/datatransfer/v1"
	"github.com/pkg/errors"
)

//WorkloadDownload command
type WorkloadDownload struct {
	*command
}

//WorkloadDownloadFactory returns a factory method for the join command
func WorkloadDownloadFactory() (cli.Command, error) {
	comm, err := newCommand("nerd workload download <workload-id> <output-dir>", "Download output data of a workload.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &WorkloadDownload{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkloadDownload) DoRun(args []string) (err error) {
	if len(args) < 2 {
		return errors.Wrap(errShowHelp, "Not enough arguments, see below for usage.")
	}

	workloadID := args[0]
	outputDir := args[1]

	// Folder create and check
	fi, err := os.Stat(outputDir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, OutputDirPermissions)
		if err != nil {
			return HandleError(errors.Errorf("The provided path '%s' does not exist and could not be created.", outputDir))
		}
		fi, err = os.Stat(outputDir)
	}
	if err != nil {
		return HandleError(err)
	} else if !fi.IsDir() {
		return HandleError(errors.Errorf("The provided path '%s' is not a directory", outputDir))
	}

	// Clients
	batchclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}
	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}

	projectID, err := ss.RequireProjectID()
	if err != nil {
		return HandleError(err)
	}

	dataOps, err := aws.NewDataClient(
		aws.NewNerdalizeCredentials(batchclient, projectID),
		ss.Project.AWSRegion,
	)
	if err != nil {
		return HandleError(errors.Wrap(err, "could not create aws dataops client"))
	}

	// Gather dataset IDs
	tasks, err := batchclient.ListTasks(ss.Project.Name, workloadID, true)
	if err != nil {
		return HandleError(err)
	}

	for _, task := range tasks.Tasks {
		if task.OutputDatasetID == "" {
			continue
		}
		cmdString := strings.Join(task.Cmd, "")
		taskDir := fmt.Sprintf("%x_%v", md5.Sum([]byte(cmdString)), task.TaskID)
		localDir := path.Join(outputDir, taskDir)
		err := os.Mkdir(localDir, OutputDirPermissions)
		if os.IsExist(err) {
			cmd.outputter.Logger.Printf("Dataset %v for task %v already exists\n", task.OutputDatasetID, task.TaskID)
			continue
		}
		downloadConf := v1datatransfer.DownloadConfig{
			BatchClient: batchclient,
			DataOps:     dataOps,
			LocalDir:    localDir,
			ProjectID:   ss.Project.Name,
			DatasetID:   task.OutputDatasetID,
			Concurrency: DownloadConcurrency,
		}
		cmd.outputter.Logger.Printf("Downloading dataset with ID '%v'", task.OutputDatasetID)
		progressCh := make(chan int64)
		progressBarDoneCh := make(chan struct{})
		var size int64
		size, err = v1datatransfer.GetRemoteDatasetSize(context.Background(), batchclient, dataOps, ss.Project.Name, task.OutputDatasetID)
		if err != nil {
			return HandleError(err)
		}
		go ProgressBar(cmd.outputter.ErrW(), size, progressCh, progressBarDoneCh)
		downloadConf.ProgressCh = progressCh
		err = v1datatransfer.Download(context.Background(), downloadConf)
		if err != nil {
			return HandleError(errors.Wrapf(err, "failed to download dataset '%v'", task.OutputDatasetID))
		}
		<-progressBarDoneCh
	}

	return nil
}

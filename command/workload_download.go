package command

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
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
	comm, err := newCommand("nerd workload download <workload-id> <output-dir>", "download output data of a workload", "", nil)
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
		return fmt.Errorf("not enough arguments, see --help")
	}

	workloadID := args[0]
	outputDir := args[1]

	// Folder create and check
	fi, err := os.Stat(outputDir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, OutputDirPermissions)
		if err != nil {
			HandleError(errors.Errorf("The provided path '%s' does not exist and could not be created.", outputDir))
		}
		fi, err = os.Stat(outputDir)
	}
	if err != nil {
		HandleError(err)
	} else if !fi.IsDir() {
		HandleError(errors.Errorf("The provided path '%s' is not a directory", outputDir))
	}

	// Clients
	batchclient, err := NewClient(cmd.config, cmd.session)
	if err != nil {
		HandleError(err)
	}
	ss, err := cmd.session.Read()
	if err != nil {
		HandleError(err)
	}
	dataOps, err := aws.NewDataClient(
		aws.NewNerdalizeCredentials(batchclient, ss.Project.Name),
		ss.Project.AWSRegion,
	)
	if err != nil {
		HandleError(errors.Wrap(err, "could not create aws dataops client"))
	}

	// Gather dataset IDs
	tasks, err := batchclient.ListTasks(ss.Project.Name, workloadID, true)
	if err != nil {
		HandleError(err)
	}

	for _, task := range tasks.Tasks {
		cmdString := strings.Join(task.Cmd, "")
		taskDir := fmt.Sprintf("%x", md5.Sum([]byte(cmdString)))
		localDir := path.Join(outputDir, taskDir)
		err := os.Mkdir(localDir, OutputDirPermissions)
		if os.IsExist(err) {
			logrus.Infof("Dataset %v for task %v already exists\n", task.OutputDatasetID, task.TaskID)
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
		logrus.Infof("Downloading dataset with ID '%v'", task.OutputDatasetID)
		progressCh := make(chan int64)
		progressBarDoneCh := make(chan struct{})
		var size int64
		size, err = v1datatransfer.GetRemoteDatasetSize(context.Background(), batchclient, dataOps, ss.Project.Name, task.OutputDatasetID)
		if err != nil {
			HandleError(err)
		}
		go ProgressBar(size, progressCh, progressBarDoneCh)
		downloadConf.ProgressCh = progressCh
		err = v1datatransfer.Download(context.Background(), downloadConf)
		if err != nil {
			HandleError(errors.Wrapf(err, "failed to download dataset '%v'", task.OutputDatasetID))
		}
	}

	return nil
}

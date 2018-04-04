package command

import (
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/cli"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/nerdalize/nerd/cmd"
	v1payload "github.com/nerdalize/nerd/nerd/client/auth/v1/payload"

	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/nerdalize/nerd/pkg/populator"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

//ProjectSetOpts determine
type ProjectSetOpts struct {
	Config     string `long:"config-src" default:"oidc" default-mask:"" description:"type of configuration to use (from env, endpoint, or oidc)"`
	KubeConfig string `long:"kubeconfig" env:"KUBECONFIG" description:"file at which Nerd will look for Kubernetes credentials" default-mask:"~/.kube/config"`
}

//ProjectSet command
type ProjectSet struct {
	*command
	opts *ProjectSetOpts
}

//ProjectSetFactory returns a factory method for the join command
func ProjectSetFactory() (cli.Command, error) {
	opts := &ProjectSetOpts{}
	comm, err := newCommand("nerd project set", "Set current working project.", "", opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &ProjectSet{
		command: comm,
		opts:    opts,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *ProjectSet) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return errShowHelp("Not enough arguments, see below for usage.")
	}
	projectSlug := args[0]

	err = cmd.session.WriteProject(projectSlug, conf.DefaultAWSRegion)
	if err != nil {
		return HandleError(err)
	}

	cmd.outputter.Logger.Printf("Project %s set successfully", projectSlug)
	return nil
}

func setProject(c *populator.Client, kubeConfig, conf string, project *v1payload.GetProjectOutput, logger *log.Logger) error {
	var (
		hdir string
		err  error
	)
	hdir, err = homedir.Dir()
	if err != nil {
		return err
	}
	if kubeConfig == "" {
		kubeConfig = filepath.Join(hdir, ".kube", "config")
	}
	p, err := populator.New(c, conf, kubeConfig, hdir, project)
	if err != nil {
		return err
	}
	err = p.PopulateKubeConfig(project.Nk)
	if err != nil {
		p.RemoveConfig(project.Nk)
		return err
	}
	if err := checkNamespace(kubeConfig, project.Nk); err != nil {
		p.RemoveConfig(project.Nk)
		return err
	}
	return nil

}

func checkNamespace(kubeConfig, ns string) error {
	kcfg, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		if os.IsNotExist(err) {
			return cmd.ErrNotLoggedIn
		}
		return errors.Wrap(err, "failed to build Kubernetes config from provided kube config path")
	}
	kube, err := kubernetes.NewForConfig(kcfg)
	if err != nil {
		return errors.Wrap(err, "failed to create Kubernetes configuration")
	}

	_, err = kube.CoreV1().Namespaces().Get(ns, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return nil
}

package transfer

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
)

//kubeDelegate updates metadat in kubernetes after lifecycle events
type kubeDelegate struct {
	name string
	kube *svc.Kube
}

func (d *kubeDelegate) PostClean(ctx context.Context) error {
	return d.PostPush(ctx, 0)
}

func (d *kubeDelegate) PostPush(ctx context.Context, size uint64) error {
	if _, err := d.kube.UpdateDataset(ctx, &svc.UpdateDatasetInput{
		Name: d.name,
		Size: &size,
	}); err != nil {
		return errors.Wrap(err, "failed to update dataset")
	}

	return nil
}

func (d *kubeDelegate) PostPull(ctx context.Context) error { return nil }
func (d *kubeDelegate) PostClose() error                   { return nil }

//KubeManager is a dataset manager that uses Kubernetes as its metadata
//store and locking service
type KubeManager struct {
	kube *svc.Kube
}

//NewKubeManager creates a transferManager that uses our kubevisor implementation
func NewKubeManager(kube *svc.Kube) (mgr *KubeManager, err error) {
	mgr = &KubeManager{
		kube: kube,
	}

	return mgr, nil
}

//Create a dataset with provided name and return a handle to it, dataset must not yet exist
func (mgr *KubeManager) Create(ctx context.Context, name string, st StoreType, at ArchiverType, opts map[string]string) (h Handle, err error) {

	//step 0: implementation options for
	if opts == nil {
		opts = map[string]string{}
	}

	d := make([]byte, 16)
	_, err = rand.Read(d)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read random bytes")
	}

	//archiver is in control of key prefixes inside the store prefix
	//@TODO we would preferrably have the kubernetes name here as well
	//but that one can be generated and is options, hence is only known
	//after the dataset has been created
	//@TODO this should probably a mandatory argument of any archiver so
	//to be addedd to the ArchiverFactory type
	opts["tar_key_prefix"] = fmt.Sprintf("%x/", d)

	//step 1: initate stores and archivers from options
	store, err := CreateStore(st, opts)
	if err != nil {
		return nil, errors.Errorf("failed to setup store '%s' with options: %#v", st, opts)
	}

	archiver, err := CreateArchiver(at, opts)
	if err != nil {
		return nil, errors.Errorf("failed to setup archiver '%s' with options: %#v", at, opts)
	}

	//step 3: create the dataset resource
	in := &svc.CreateDatasetInput{
		Name:         name,
		Size:         0,
		StoreType:    string(st),
		ArchiverType: string(at),
		Options:      opts,
	}

	out, err := mgr.kube.CreateDataset(ctx, in)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create dataset resource")
	}

	//step 2: initiate the handle
	return CreateStdHandle(out.Name, store, archiver, &kubeDelegate{
		name: out.Name,
		kube: mgr.kube,
	})
}

//Open an existing dataset and return a handle to it, dataset must exist
func (mgr *KubeManager) Open(ctx context.Context, name string) (Handle, error) {
	in := &svc.GetDatasetInput{
		Name: name,
	}

	out, err := mgr.kube.GetDataset(ctx, in)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get dataset resource")
	}

	store, err := CreateStore(StoreType(out.StoreType), out.Options)
	if err != nil {
		return nil, errors.Errorf("failed to setup store '%s' with options: %#v", out.StoreType, out.Options)
	}

	archiver, err := CreateArchiver(ArchiverType(out.ArchiverType), out.Options)
	if err != nil {
		return nil, errors.Errorf("failed to setup archiver '%s' with options: %#v", out.ArchiverType, out.Options)
	}

	return CreateStdHandle(out.Name, store, archiver, &kubeDelegate{
		name: out.Name,
		kube: mgr.kube,
	})
}

//Remove an existing dataset, dataset must exist
func (mgr *KubeManager) Remove(ctx context.Context, name string) error {
	_, err := mgr.kube.DeleteDataset(ctx, &svc.DeleteDatasetInput{Name: name})
	if err != nil {
		return errors.Wrap(err, "failed to delete resource")
	}

	return nil
}

//Info (re)fetches dataset info from the manager
func (mgr *KubeManager) Info(ctx context.Context, name string) (size uint64, err error) {
	out, err := mgr.kube.GetDataset(ctx, &svc.GetDatasetInput{
		Name: name,
	})

	if err != nil {
		return 0, errors.Wrap(err, "failed to get dataset resource")
	}

	return out.Size, nil
}

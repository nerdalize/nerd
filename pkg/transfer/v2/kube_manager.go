package transfer

import (
	"context"

	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
)

type kubeHandle struct {
	kube *svc.Kube

	store    Store
	archiver interface{} //@TODO what interface do we need here?
}

func (h *kubeHandle) Upload(fromPath string, progress chan<- struct{}) error {
	return nil
}

func (h *kubeHandle) Download(toPath string, progress chan<- struct{}) error {
	return nil
}

func (h *kubeHandle) Close() error {
	return nil
}

//KubeManager is a dataset manager that uses Kubernetes as its metadata
//store and locking service
type KubeManager struct {
	kube      *svc.Kube
	stores    map[StoreType]Store
	archivers map[ArchiverType]interface{}
}

//NewKubeManager creates a transferManager that uses our kubevisor implementation
func NewKubeManager(kube *svc.Kube, stores map[StoreType]Store, archivers map[ArchiverType]interface{}) (mgr *KubeManager, err error) {
	mgr = &KubeManager{kube: kube, stores: stores, archivers: archivers}
	return mgr, nil
}

//Create a dataset with provided name and return a handle to it, dataset must not yet exist
func (mgr *KubeManager) Create(ctx context.Context, name string, st StoreType, at ArchiverType) (h Handle, err error) {
	store, ok := mgr.stores[st]
	if !ok {
		return nil, errors.Errorf("store type '%s' is not configured in the manager", st)
	}

	archiver, ok := mgr.archivers[at]
	if !ok {
		return nil, errors.Errorf("archiver '%s' is not configured in the manager", at)
	}

	//step 1: initiate the handle
	h = &kubeHandle{
		kube:     mgr.kube,
		store:    store,
		archiver: archiver,
	}

	//step 2: create the dataset resource
	in := &svc.CreateDatasetInput{
		//@TODO store store type
		//@TODO store arhive type
		Name:   "datasetName",
		Bucket: "ref.Bucket",
		Key:    "ref.Key",
		// Size:   uint64(n),
	}

	out, err := mgr.kube.CreateDataset(ctx, in)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create dataset resource")
	}

	//step 3: fill handle with resource info
	_ = out

	return h, nil
}

//Open an existing dataset and return a handle to it, dataset must exist
func (mgr *KubeManager) Open(ctx context.Context, name string) (Handle, error) {
	return nil, nil
}

//Remove an existing dataset, dataset must exist
func (mgr *KubeManager) Remove(ctx context.Context, name string) error {
	return nil
}

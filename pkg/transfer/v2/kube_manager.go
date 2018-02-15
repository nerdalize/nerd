package transfer

import (
	"context"

	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
)

type kubeMeta struct {
	name string
	kube *svc.Kube
}

//Name returns the unique handle name
func (h *kubeMeta) Name() string {
	return h.name
}

func (h *kubeMeta) Close() error {
	//@TODO remove lock, for now a no-op
	return nil
}

func (h *kubeMeta) UpdateMeta(ctx context.Context, size uint64) error {
	if _, err := h.kube.UpdateDataset(ctx, &svc.UpdateDatasetInput{
		Name: h.name,
		Size: &size,
	}); err != nil {
		return errors.Wrap(err, "failed to update dataset")
	}

	return nil
}

func (h *kubeMeta) ReadMeta(ctx context.Context) (size uint64, err error) {
	out, err := h.kube.GetDataset(ctx, &svc.GetDatasetInput{Name: h.name})
	if err != nil {
		return 0, errors.Wrap(err, "failed to get dataset")
	}

	return out.Size, nil
}

//KubeManager is a dataset manager that uses Kubernetes as its metadata
//store and locking service
type KubeManager struct {
	kube      *svc.Kube
	stores    map[StoreType]StoreFactory
	archivers map[ArchiverType]ArchiverFactory
}

//NewKubeManager creates a transferManager that uses our kubevisor implementation
func NewKubeManager(kube *svc.Kube, stores map[StoreType]StoreFactory, archivers map[ArchiverType]ArchiverFactory) (mgr *KubeManager, err error) {
	mgr = &KubeManager{
		kube:      kube,
		stores:    stores,
		archivers: archivers,
	}

	return mgr, nil
}

//Create a dataset with provided name and return a handle to it, dataset must not yet exist
func (mgr *KubeManager) Create(ctx context.Context, name string, st StoreType, at ArchiverType, opts map[string]string) (h Handle, err error) {

	//step 1: initate stores and archivers from options
	storef, ok := mgr.stores[st]
	if !ok {
		return nil, errors.Errorf("store type '%s' is not configured in the manager", st)
	}

	store, err := storef(opts)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create store from options")
	}

	archiverf, ok := mgr.archivers[at]
	if !ok {
		return nil, errors.Errorf("archiver '%s' is not configured in the manager", at)
	}

	archiver, err := archiverf(opts)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create archiver from options")
	}

	//step 2: initiate the handle
	kh := &StdHandle{
		store:    store,
		archiver: archiver,
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

	//step 3: fill handle with resource info
	kh.Meta = &kubeMeta{
		name: out.Name,
		kube: mgr.kube,
	}

	return kh, nil
}

//Open an existing dataset and return a handle to it, dataset must exist
func (mgr *KubeManager) Open(ctx context.Context, name string) (Handle, error) {

	//Step 1: Get the dataset by name
	in := &svc.GetDatasetInput{
		Name: name,
	}

	out, err := mgr.kube.GetDataset(ctx, in)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get dataset resource")
	}

	//Step 2: Create store and archivers
	storef, ok := mgr.stores[StoreType(out.StoreType)]
	if !ok {
		return nil, errors.Errorf("store type '%s' is not configured in the manager", out.StoreType)
	}

	store, err := storef(out.Options)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create store from options")
	}

	archiverf, ok := mgr.archivers[ArchiverType(out.ArchiverType)]
	if !ok {
		return nil, errors.Errorf("archiver '%s' is not configured in the manager", out.ArchiverType)
	}

	archiver, err := archiverf(out.Options)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create archiver from options")
	}

	//step 3: Create handle
	kh := &StdHandle{
		store:    store,
		archiver: archiver,

		Meta: &kubeMeta{
			name: out.Name,
			kube: mgr.kube,
		},
	}

	return kh, nil
}

//Remove an existing dataset, dataset must exist
func (mgr *KubeManager) Remove(ctx context.Context, name string) error {
	_, err := mgr.kube.DeleteDataset(ctx, &svc.DeleteDatasetInput{Name: name})
	if err != nil {
		return errors.Wrap(err, "failed to delete resource")
	}

	return nil
}

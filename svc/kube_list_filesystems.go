package svc

import (
	"context"
	"time"

	"github.com/nerdalize/nerd/pkg/kubevisor"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

//FileSystemDetails tells us more about the filesystem by looking at underlying resources
type FileSystemDetails struct {
	CreatedAt  time.Time
	Size       uint64
	Status     v1.PersistentVolumeClaimPhase
}

//ListFileSystemItem is a filesystem listing item
type ListFileSystemItem struct {
	Name    string
	Details FileSystemDetails
}

//ListFileSystemsInput is the input to ListFileSystems
type ListFileSystemsInput struct{}

//ListFileSystemsOutput is the output to ListFileSystems
type ListFileSystemsOutput struct {
	Items []*ListFileSystemItem
}

//ListFileSystems will create a filesystem on kubernetes
func (k *Kube) ListFileSystems(ctx context.Context, in *ListFileSystemsInput) (out *ListFileSystemsOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	//Step 0: Get all the filesystems under nerd-app=cli
	filesystems := &filesystems{}
	err = k.visor.ListResources(ctx, kubevisor.ResourceTypePersistentVolumeClaims, filesystems, nil, nil)
	if err != nil {
		return nil, err
	}

	//Step 1: Analyse filesystem structure and formulate our output items
	out = &ListFileSystemsOutput{}
	mapping := map[types.UID]*ListFileSystemItem{}
	for _, filesystem := range filesystems.Items {
		capacity := filesystem.Status.Capacity[v1.ResourceName(v1.ResourceStorage)]
                size64, _ := capacity.AsInt64()

		item := &ListFileSystemItem{
			Name: filesystem.GetName(),
			Details: FileSystemDetails{
				Size:       uint64(size64),
				CreatedAt:  filesystem.CreationTimestamp.Local(),
				Status:     filesystem.Status.Phase,
			},
		}

		mapping[filesystem.UID] = item
		out.Items = append(out.Items, item)
	}

	return out, nil
}

//filesystems implements the list transformer interface to allow the kubevisor to manage names for us
type filesystems struct{ *v1.PersistentVolumeClaimList }

func (filesystems *filesystems) Transform(fn func(in kubevisor.ManagedNames) (out kubevisor.ManagedNames)) {
	for i, d1 := range filesystems.PersistentVolumeClaimList.Items {
		filesystems.Items[i] = *(fn(&d1).(*v1.PersistentVolumeClaim))
	}
}

func (filesystems *filesystems) Len() int {
	return len(filesystems.PersistentVolumeClaimList.Items)
}

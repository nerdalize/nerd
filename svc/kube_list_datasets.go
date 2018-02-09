package svc

import (
	"context"
	"time"

	datasetsv1 "github.com/nerdalize/nerd/crd/pkg/apis/stable.nerdalize.com/v1"
	"github.com/nerdalize/nerd/pkg/kubevisor"

	"k8s.io/apimachinery/pkg/types"
)

//DatasetDetails tells us more about the dataset by looking at underlying resources
type DatasetDetails struct {
	CreatedAt  time.Time
	Size       uint64
	InputFor   []string
	OutputFrom []string
}

//ListDatasetItem is a dataset listing item
type ListDatasetItem struct {
	Name    string
	Details DatasetDetails
}

//ListDatasetsInput is the input to ListDatasets
type ListDatasetsInput struct{}

//ListDatasetsOutput is the output to ListDatasets
type ListDatasetsOutput struct {
	Items []*ListDatasetItem
}

//ListDatasets will create a dataset on kubernetes
func (k *Kube) ListDatasets(ctx context.Context, in *ListDatasetsInput) (out *ListDatasetsOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	//Step 0: Get all the datasets under nerd-app=cli
	datasets := &datasets{}
	err = k.visor.ListResources(ctx, kubevisor.ResourceTypeDatasets, datasets, nil)
	if err != nil {
		return nil, err
	}

	//Step 1: Analyse dataset structure and formulate our output items
	out = &ListDatasetsOutput{}
	mapping := map[types.UID]*ListDatasetItem{}
	for _, dataset := range datasets.Items {
		item := &ListDatasetItem{
			Name: dataset.GetName(),
			Details: DatasetDetails{
				Size:       dataset.Spec.Size,
				InputFor:   dataset.Spec.InputFor,
				OutputFrom: dataset.Spec.OutputFrom,
				CreatedAt:  dataset.CreationTimestamp.Local(),
			},
		}

		mapping[dataset.UID] = item
		out.Items = append(out.Items, item)
	}

	return out, nil
}

//datasets implements the list transformer interface to allow the kubevisor to manage names for us
type datasets struct{ *datasetsv1.DatasetList }

func (datasets *datasets) Transform(fn func(in kubevisor.ManagedNames) (out kubevisor.ManagedNames)) {
	for i, d1 := range datasets.DatasetList.Items {
		datasets.Items[i] = *(fn(&d1).(*datasetsv1.Dataset))
	}
}

func (datasets *datasets) Len() int {
	return len(datasets.DatasetList.Items)
}

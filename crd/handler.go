package main

import (
	"context"

	"github.com/golang/glog"
	datasetsv1 "github.com/nerdalize/nerd/crd/pkg/apis/stable.nerdalize.com/v1"
	transferv2 "github.com/nerdalize/nerd/pkg/transfer"
)

type glogReporter struct{}

func (r *glogReporter) HandledKey(key string) {
	glog.Infof("handled dataset key '%s'", key)
}

// Handler is implemented by any handler.
// The Handle method is used to process event
type Handler interface {
	ObjectCreated(obj interface{})
	ObjectDeleted(obj interface{}, key string)
	ObjectUpdated(oldObj, newObj interface{})
}

// S3AWS handler implements Handler interface
type S3AWS struct{}

// ObjectCreated will be called each time an object is created
func (s *S3AWS) ObjectCreated(obj interface{}) {
	if dataset, ok := obj.(*datasetsv1.Dataset); ok {
		glog.Infof("New dataset created %s from namespace %s", dataset.Name, dataset.Namespace)
	}
}

// ObjectDeleted will be called each time an object is deleted
// If the object is a dataset, the corresponding dataset will be removed from s3
func (s *S3AWS) ObjectDeleted(obj interface{}, key string) {
	if dataset, ok := obj.(*datasetsv1.Dataset); ok {
		store, err := transferv2.CreateStore(dataset.Spec.StoreOptions)
		if err != nil {
			glog.Errorf("failed to create store with type '%s': %v", dataset.Spec.StoreType, err)
			return
		}

		archiver, err := transferv2.CreateArchiver(dataset.Spec.ArchiverOptions)
		if err != nil {
			glog.Errorf("failed to create archiver with type '%s': %v", dataset.Spec.ArchiverType, err)
			return
		}

		h, err := transferv2.CreateStdHandle(dataset.GetName(), store, archiver, nil)
		if err != nil {
			glog.Errorf("failed to create standard handle: %v", err)
			return
		}

		//@TODO decide on the timeout of the dataset clear
		err = h.Clear(context.TODO(), &glogReporter{})
		if err != nil {
			glog.Errorf("failed to clear the dataset: %v", err)
			return
		}

		glog.Infof("Dataset deleted %s from namespace %s", dataset.Name, dataset.Namespace)
	}
}

// ObjectUpdated will be called each time an object is updated
func (s *S3AWS) ObjectUpdated(oldObj, newObj interface{}) {
	if dataset, ok := newObj.(*datasetsv1.Dataset); ok {
		glog.Infof("New dataset updated %s from namespace %s", dataset.Name, dataset.Namespace)
	}
}

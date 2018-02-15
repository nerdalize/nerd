package main

import (
	"context"

	"github.com/golang/glog"
	datasetsv1 "github.com/nerdalize/nerd/crd/pkg/apis/stable.nerdalize.com/v1"
	transfer "github.com/nerdalize/nerd/pkg/transfer/v1"
	v2transfer "github.com/nerdalize/nerd/pkg/transfer/v2"
)

// Handler is implemented by any handler.
// The Handle method is used to process event
type Handler interface {
	Init() error
	ObjectCreated(obj interface{})
	ObjectDeleted(obj interface{}, key string)
	ObjectUpdated(oldObj, newObj interface{})
}

// S3AWS handler implements Handler interface
type S3AWS struct {
	conf *transfer.S3
}

// Init instantiates an aws s3 client that we'll use to manage datasets
func (s *S3AWS) Init() error {

	//@TODO use the transfer/v2 implementation
	s3cfg := &transfer.S3Conf{
		Bucket: "nlz-datasets-dev",
	}

	s3, err := transfer.NewS3(s3cfg)
	if err != nil {
		return nil
	}

	s.conf = s3
	return nil
}

// ObjectCreated will be called each time an object is created
func (s *S3AWS) ObjectCreated(obj interface{}) {
	if dataset, ok := obj.(*datasetsv1.Dataset); ok {
		glog.Infof("New dataset created: %s with bucket %s and key %s", dataset.Name, dataset.Spec.Bucket, dataset.Spec.Key)
	}
}

// ObjectDeleted will be called each time an object is deleted
// If the object is a dataset, the corresponding dataset will be removed from s3
func (s *S3AWS) ObjectDeleted(obj interface{}, key string) {

	//@TODO we emulate transfer/v2 format to create a v1 ref for deletion
	var ref transfer.Ref
	if dataset, ok := obj.(*datasetsv1.Dataset); ok {
		opts := dataset.Spec.Options
		if opts != nil {
			ref.Bucket = opts["aws_s3_bucket"]
			ref.Key = v2transfer.TarArchiverKey
		}
	} else {
		glog.Infof("Object %s not found %+v", key, obj)
		return
	}
	if err := s.conf.Delete(context.Background(), &ref); err != nil {
		glog.Errorf("Could not delete dataset from aws: %+v", err)
	}

	glog.Infof("New dataset deleted: %s with bucket %s and key %s", key, ref.Bucket, ref.Key)
}

// ObjectUpdated will be called each time an object is updated
func (s *S3AWS) ObjectUpdated(oldObj, newObj interface{}) {
	if dataset, ok := newObj.(*datasetsv1.Dataset); ok {
		glog.Infof("New dataset updated: %s with bucket %s and key %s", dataset.Name, dataset.Spec.Bucket, dataset.Spec.Key)
	}

}

/*
Copyright 2018 The Openshift Evangelists

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package v1

import (
	v1 "github.com/nerdalize/nerd/crd/pkg/apis/nerdalize.com/v1"
	scheme "github.com/nerdalize/nerd/crd/pkg/client/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// DatasetsGetter has a method to return a DatasetInterface.
// A group's client should implement this interface.
type DatasetsGetter interface {
	Datasets(namespace string) DatasetInterface
}

// DatasetInterface has methods to work with Dataset resources.
type DatasetInterface interface {
	Create(*v1.Dataset) (*v1.Dataset, error)
	Update(*v1.Dataset) (*v1.Dataset, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.Dataset, error)
	List(opts meta_v1.ListOptions) (*v1.DatasetList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Dataset, err error)
	DatasetExpansion
}

// datasets implements DatasetInterface
type datasets struct {
	client rest.Interface
	ns     string
}

// newDatasets returns a Datasets
func newDatasets(c *NerdalizeV1Client, namespace string) *datasets {
	return &datasets{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the dataset, and returns the corresponding dataset object, and an error if there is any.
func (c *datasets) Get(name string, options meta_v1.GetOptions) (result *v1.Dataset, err error) {
	result = &v1.Dataset{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("datasets").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Datasets that match those selectors.
func (c *datasets) List(opts meta_v1.ListOptions) (result *v1.DatasetList, err error) {
	result = &v1.DatasetList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("datasets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested datasets.
func (c *datasets) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("datasets").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a dataset and creates it.  Returns the server's representation of the dataset, and an error, if there is any.
func (c *datasets) Create(dataset *v1.Dataset) (result *v1.Dataset, err error) {
	result = &v1.Dataset{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("datasets").
		Body(dataset).
		Do().
		Into(result)
	return
}

// Update takes the representation of a dataset and updates it. Returns the server's representation of the dataset, and an error, if there is any.
func (c *datasets) Update(dataset *v1.Dataset) (result *v1.Dataset, err error) {
	result = &v1.Dataset{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("datasets").
		Name(dataset.Name).
		Body(dataset).
		Do().
		Into(result)
	return
}

// Delete takes name of the dataset and deletes it. Returns an error if one occurs.
func (c *datasets) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("datasets").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *datasets) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("datasets").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched dataset.
func (c *datasets) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Dataset, err error) {
	result = &v1.Dataset{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("datasets").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}

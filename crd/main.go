/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"flag"
	"time"

	"github.com/golang/glog"
	"k8s.io/client-go/tools/clientcmd"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	clientset "github.com/nerdalize/nerd/crd/pkg/client/clientset/versioned"
	informers "github.com/nerdalize/nerd/crd/pkg/client/informers/externalversions"
	"github.com/nerdalize/nerd/crd/pkg/signals"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	glog.Info("Setting up signal handler")
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	datasetClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building dataset clientset: %s", err.Error())
	}

	datasetInformerFactory := informers.NewSharedInformerFactory(datasetClient, time.Second*30)
	eventHandler := new(S3AWS)
	err = eventHandler.Init()
	if err != nil {
		glog.Fatalf("Error while init event handler : %+v", err)
	}
	controller := NewController(datasetClient, datasetInformerFactory, eventHandler)

	go datasetInformerFactory.Start(stopCh)

	controller.Run(stopCh)
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

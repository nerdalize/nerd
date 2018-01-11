package main

import (
	"flag"
	"fmt"

	"github.com/golang/glog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	nerdalizecs "github.com/nerdalize/nerd/crd/pkg/client/clientset/versioned"
)

var (
	kuberconfig = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	master      = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
)

func main() {
	flag.Parse()

	cfg, err := clientcmd.BuildConfigFromFlags(*master, *kuberconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	exampleClient, err := nerdalizecs.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building example clientset: %v", err)
	}

	list, err := exampleClient.NerdalizeV1().Datasets("default").List(metav1.ListOptions{})
	if err != nil {
		glog.Fatalf("Error listing all datasets: %v", err)
	}

	for _, db := range list.Items {
		fmt.Printf("datasets %s with key %q and bucket %q\n", db.Name, db.Spec.Key, db.Spec.Bucket)
	}
}

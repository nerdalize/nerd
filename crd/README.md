# Custom Dataset controller

This controller should be deployed on each cluster.
It is used to delete S3 object when a Kubernetes Dataset is being deleted.

## Some resources regarding controllers

- [Writing custom Kubernetes controllers](https://medium.com/@cloudark/kubernetes-custom-controllers-b6c7d0668fdf), it's a must read and there is a very good representation of how a custom controller works.

- [Kubernetes sample controller](https://github.com/kubernetes/sample-controller)

- [Kubewatch, an example of Kubernetes Custom Controller](https://engineering.bitnami.com/articles/kubewatch-an-example-of-kubernetes-custom-controller.html)

## How it works

![interaction between a custom controller and client-go](https://cdn-images-1.medium.com/max/800/1*dmvNSeSIORAMaTF2WdE9fg.jpeg)
*Illustration from [Writing custom Kubernetes controllers](https://medium.com/@cloudark/kubernetes-custom-controllers-b6c7d0668fdf)*

- The controller creates an Informer and an Indexer to list, watch and index a Kubernetes object, for our controller it's all about datasets

- When a new object is being deleted, it calls an event handler to delete the s3 object.

## Running it locally

```bash
$ go run *.go -kubeconfig=</PATH/TO/YOUR/KUBECONFIG> -alsologtostderr -v 4
```

## Building the docker image

```bash
$ CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o controller .
$ docker build -t nerdalize/custom-dataset-controller:<TAG> -f Dockerfile .
$ docker push nerdalize/custom-dataset-controller
```

## Deploying the controller on Kubernetes

Update the docker image tag in the [deployment file](https://github.com/nerdalize/nerd/blob/master/crd/deployment.yml) and then apply it using kubectl:

```bash
$ kubectl apply -f deployment.yml
```

The deployment is always done in the `kube-system` namespace.
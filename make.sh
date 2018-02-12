#!/bin/bash
set -e

dev_profile="nerd-cli-dev"

function print_help {
	printf "Available Commands:\n";
	awk -v sq="'" '/^function run_([a-zA-Z0-9-]*)\s*/ {print "-e " sq NR "p" sq " -e " sq NR-1 "p" sq }' make.sh \
		| while read line; do eval "sed -n $line make.sh"; done \
		| paste -d"|" - - \
		| sed -e 's/^/  /' -e 's/function run_//' -e 's/#//' -e 's/{/	/' \
		| awk -F '|' '{ print "  " $2 "\t" $1}' \
		| expand -t 30
}

function run_build { #compile versioned executable and place it in $GOPATH/bin
	go build \
    -ldflags "-X main.version=$(cat VERSION) -X main.commit=$(git rev-parse --short HEAD )" \
	-tags "forceposix" \
    -o $GOPATH/bin/nerd \
    main.go
}

function run_dev { #setup dev environment
	command -v go >/dev/null 2>&1 || { echo "executable 'go' (the language sdk) must be installed" >&2; exit 1; }
	command -v minikube >/dev/null 2>&1 || { echo "executable 'minikube' (local kubernetes cluster) must be installed" >&2; exit 1; }
	command -v kubectl >/dev/null 2>&1 || { echo "executable 'kubectl' (kubernetes cli https://kubernetes.io/docs/tasks/tools/install-kubectl/) must be installed" >&2; exit 1; }
	command -v glide >/dev/null 2>&1 || { echo "executable glide (https://github.com/Masterminds/glide) must be installed" >&2; exit 1; }

	#develop against specific version and configure flex volume to reflect prod setup
	kube_version="v1.8.0"
	flexvolume_config="--extra-config=controller-manager.FlexVolumePluginDir=/var/lib/kubelet/volumeplugins/ --extra-config=kubelet.VolumePluginDir=/var/lib/kubelet/volumeplugins/"
	if minikube status --profile=$dev_profile | grep Running; then
	    echo "--> minikube vm (profile: $dev_profile) is already running (check: $kube_version), skipping restart"
			minikube profile $dev_profile
	else
			echo "--> starting minikube using the default 'vm-driver',to configure: https://github.com/kubernetes/minikube/issues/637)"
		  minikube start $flexvolume_config --profile=$dev_profile --kubernetes-version=$kube_version

			echo "--> sleeping to let k8s initial setup take place"
			sleep 10
	fi

	echo "--> setting up kube config"
	kubectl config set-context $dev_profile --user=$dev_profile --cluster=$dev_profile --namespace=default && kubectl config use-context $dev_profile

	echo "--> setting up custom resource definition for datasets"
	kubectl apply -f crd/artifacts/datasets.yaml

	echo "--> installing flex volume deamon set"
	kubectl apply -f cmd/flex/dataset.yml

	echo "--> updating dependencies"
	glide up

	echo "--> checking crd generated code is valid"
	if ./crd/hack/verify-codegen.sh; then
		echo "--> crd code is up-to-date"
	else
		echo "--> regenerating code for crd"
		./crd/hack/update-codegen.sh
	fi
}

function run_docs { #run godoc
  command -v go >/dev/null 2>&1 || { echo "executable 'go' (the language sdk) must be installed" >&2; exit 1; }

	echo "--> starting godoc service (http://localhost:6060/pkg/github.com/nerdalize/nerd)"
	godoc -v -http=":6060"
}

function run_test { #unit test project
	command -v go >/dev/null 2>&1 || { echo "executable 'go' (the language sdk) must be installed" >&2; exit 1; }

	echo "--> running service tests"
	go test -cover -v ./svc/...
}

function run_release { #cross compile new release builds
	mkdir -p bin
	gox -ldflags "-X main.version=$(cat VERSION) -X main.commit=$(git rev-parse --short HEAD )" -tags "forceposix" -osarch="linux/amd64 windows/amd64 darwin/amd64" -output=./bin/{{.OS}}_{{.Arch}}/nerd
}

function run_publish { #publish cross compiled binaries
	cd bin/darwin_amd64; tar -zcvf ../nerd-$(cat ../../VERSION)-macos.tar.gz nerd
	cd ../linux_amd64; tar -zcvf ../nerd-$(cat ../../VERSION)-linux.tar.gz nerd
	cd ../windows_amd64; zip ../nerd-$(cat ../../VERSION)-win.zip ./nerd.exe; cd ../..

	git tag v`cat VERSION` || true
	git push --tags

	github-release release \
		--user nerdalize \
		--repo nerd \
		--tag v`cat VERSION` \
		--pre-release || true

	github-release upload \
			--user nerdalize \
			--repo nerd \
			--tag v`cat VERSION` \
			--name nerd-$(cat VERSION)-macos.tar.gz \
			--file bin/nerd-$(cat VERSION)-macos.tar.gz || true

	github-release upload \
			--user nerdalize \
			--repo nerd \
			--tag v`cat VERSION` \
			--name nerd-$(cat VERSION)-linux.tar.gz \
			--file bin/nerd-$(cat VERSION)-linux.tar.gz || true

	github-release upload \
			--user nerdalize \
			--repo nerd \
			--tag v`cat VERSION` \
			--name nerd-$(cat VERSION)-win.zip \
			--file bin/nerd-$(cat VERSION)-win.zip || true
}

function run_docker { #build docker container
	command -v docker >/dev/null 2>&1 || { echo "executable 'docker' (container runtime) must be installed" >&2; exit 1; }

	echo "--> building flex volume container"
	docker build -f flex.Dockerfile -t nerdalize/nerd-flex-volume:$(cat VERSION) .
}

function run_dockerpush { #build and push docker container
	command -v docker >/dev/null 2>&1 || { echo "executable 'docker' (container runtime) must be installed" >&2; exit 1; }

	echo "--> publish flex volume container"
	docker push nerdalize/nerd-flex-volume:$(cat VERSION)
}

function run_crdbuild { #build docker container for custom dataset controller
	docker build -t nerdalize/custom-dataset-controller crd
	docker tag nerdalize/custom-dataset-controller nerdalize/custom-dataset-controller:`cat crd/VERSION`
}

function run_crdpush { #build and push docker container for custom dataset controller
	run_crdbuild
	docker push nerdalize/custom-dataset-controller:latest
	docker push nerdalize/custom-dataset-controller:`cat crd/VERSION`
}

case $1 in
	"build") run_build ;;
	"dev") run_dev ;;
	"docs") run_docs ;;
	"test") run_test ;;
	"gen") run_gen ;;
	"release") run_release ;;
	"publish") run_publish ;;
	"docker") run_docker ;;
	"dockerpush") run_dockerpush ;;
	"crdbuild") run_crdbuild ;;
	"crdpush") run_crdpush ;;
	*) print_help ;;
esac

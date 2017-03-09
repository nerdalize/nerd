#!/bin/bash
set -e

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
    -o $GOPATH/bin/nerd \
    main.go
}

function run_build-worker { #build worker container image
	docker build -t quay.io/nerdalize/worker:$(cat VERSION) -f Dockerfile.linux .
}

function run_run-worker { #run the worker container
	run_build-worker
	docker run --rm \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v ~/.nerd:/root/.nerd \
		-it quay.io/nerdalize/worker:$(cat VERSION) work
}

function run_publish-worker { #publish the worker container image
	run_build-worker
	docker push quay.io/nerdalize/worker:$(cat VERSION)
}

function run_test { #unit test project
	go test -v ./command
  go test -v ./nerd/...
}

case $1 in
	"build") run_build ;;
	"test") run_test ;;

	"build-worker") run_build-worker ;;
	"run-worker") run_run-worker ;;
	"publish-worker") run_publish-worker ;;
	*) print_help ;;
esac

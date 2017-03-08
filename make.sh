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

function run_buildgit { #build a new binary as git commit hash, and place it in $GOPATH/bin
  go build \
    -ldflags "-X main.version=$(cat VERSION) -X main.commit=$(git rev-parse --short HEAD )" \
    -o $GOPATH/bin/nerd \
    main.go
}

function run_build { #build a new binary and place it in $GOPATH/bin
  go build \
    -o $GOPATH/bin/nerd \
    main.go
}

function run_buildworker { #build the worker as a Docker container
	docker build -t quay.io/nerdalize/worker -f Dockerfile.linux .
}

function run_work { #run the worker using Docker
	docker run --rm \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v ~/.nerd:/root/.nerd \
		-it quay.io/nerdalize/worker work
}

function run_test { #unit test project
	go test -v ./command
  go test -v ./nerd/...
}

case $1 in
	"build") run_build ;;
	"buildgit") run_buildgit ;;
	"buildworker") run_buildworker ;;
	"test") run_test ;;
	"work") run_work ;;
	*) print_help ;;
esac

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

function run_buildgit { #build a new binary, tag it with git commit hash, and place it in $GOPATH/bin
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
function run_test { #unit test project
	go test -v ./command
  go test -v ./nerd/...
}

case $1 in
	"build") run_build ;;
	"buildgit") run_buildgit ;;
	"test") run_test ;;
	*) print_help ;;
esac

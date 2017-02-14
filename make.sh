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

function run_build { #build a new binary and place it in $GOPATH/bin
  go build \
    -ldflags "-X main.version=$(cat VERSION) -X main.commit=$(git rev-parse --short HEAD )" \
    -o $GOPATH/bin/nerd \
    main.go
}

function run_test_integration {
	endpoint=$(terraform output infra_endpoint)
	export TEST_NERD_API_FULL_URL=$endpoint
	export TEST_NERD_API_VERSION="v1"
	go test $(go list ./... | grep -v /vendor/) -tags=integration
}

case $1 in
	"build") run_build ;;
	"test_integration") run_test_integration ;;

	*) print_help ;;
esac

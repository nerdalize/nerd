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

function run_gen { #regenerate requesters
	command -v go >/dev/null 2>&1 || { echo "executable 'go' (the language sdk) must be installed" >&2; exit 1; }
	command -v protoc >/dev/null 2>&1 || { echo "executable 'protoc' (protobuf compiler) must be installed" >&2; exit 1; }
	command -v gg >/dev/null 2>&1 || { echo "executable 'gg' (grpc to http1.1) must be installed" >&2; exit 1; }

	echo "--> generating gRPC definitions"
	protoc \
		--go_out=plugins=grpc:$(pwd)/nerd/client/cluster								\
		--proto_path=$GOPATH/src/github.com/nerdalize/clusterd 					\
		  $GOPATH/src/github.com/nerdalize/clusterd/svc/*.proto

	echo "--> generating HTTP1.1 client from gRPC"
	gg nerd/client/cluster/svc/*.pb.go

		# @TODO how do we make sure grpc protobuf is configured write output to another directory
}

function run_test { #unit test project
	go test -v ./command/...
  go test -v ./nerd/...
}

function run_release { #cross compile new release builds
	mkdir -p bin
	gox -ldflags "-X main.version=$(cat VERSION) -X main.commit=$(git rev-parse --short HEAD )" -osarch="linux/amd64 windows/amd64 darwin/amd64" -output=./bin/{{.OS}}_{{.Arch}}/nerd
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
	docker build -t nerdalize/nerd .
	docker tag nerdalize/nerd nerdalize/nerd:`cat VERSION`
}

function run_dockerpush { #build and push docker container
	run_docker
	docker push nerdalize/nerd:latest
	docker push nerdalize/nerd:`cat VERSION`
}

case $1 in
	"build") run_build ;;
	"test") run_test ;;
	"gen") run_gen ;;
	"release") run_release ;;
	"publish") run_publish ;;
	"docker") run_docker ;;
	"dockerpush") run_dockerpush ;;
	*) print_help ;;
esac

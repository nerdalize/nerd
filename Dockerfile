FROM golang:1.8

ENV NERD_PATH /go/src/github.com/nerdalize/nerd

ADD . $NERD_PATH

RUN mkdir /in; mkdir /out
RUN cd $NERD_PATH; \
    go build \
      -ldflags "-X main.version=$(cat VERSION) -X main.commit=docker.build" \
      -o /go/bin/nerd \
      main.go

ENTRYPOINT ["/go/bin/nerd"]

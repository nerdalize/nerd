FROM golang:1-stretch as build
WORKDIR /go/src/github.com/nerdalize/nerd
COPY . .
RUN go build -o $GOPATH/bin/nerd-flex-volume cmd/flex/main.go

FROM golang:1-alpine
COPY --from=build $GOPATH/bin/nerd-flex-volume /dataset
COPY cmd/flex/install.sh /run.sh
RUN chmod +x /run.sh
ENTRYPOINT ["/run.sh"]

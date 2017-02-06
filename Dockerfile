FROM golang:latest
MAINTAINER Boris Mattijssen "b.mattijssen@nerdalize.com"

RUN apt-get update && apt-get install stress

RUN mkdir -p $GOPATH/src/github.com/nerdalize/nerd
ADD . $GOPATH/src/github.com/nerdalize/nerd
RUN cd $GOPATH/src/github.com/nerdalize/nerd; go install .

ADD entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

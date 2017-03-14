FROM golang:1.7

ENV NERD_PATH /go/src/github.com/nerdalize/nerd

ADD . $NERD_PATH

RUN mkdir /in; mkdir /out
RUN cd $NERD_PATH; ./make.sh build

CMD /go/bin/nerd

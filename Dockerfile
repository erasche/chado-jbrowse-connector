FROM golang:1.6

ADD . /go/src/github.com/erasche/jb
WORKDIR /go/src/github.com/erasche/jb
RUN make deps
RUN make
CMD ["bash", "/go/src/github.com/erasche/jb/run.sh"]

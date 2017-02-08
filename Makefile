SRC := $(wildcard *.go)
TARGET := chado-jbrowse-connector
VERSION := $(shell git describe --tags)

all: $(TARGET)

vendor: glide.yaml glide.lock
	go get github.com/Masterminds/glide/...
	go install github.com/Masterminds/glide/...
	glide install

gofmt: $(src)
	find $(SRC) -exec gofmt -w '{}' \;

qc_deps:
	go get github.com/alecthomas/gometalinter
	gometalinter --install --update

qc: lint vet complexity
	golint $(SRC)
	gocyclo -over 10 $(SRC)
	gometalinter .

test: $(SRC) vendor gofmt
	go test -v $(glide novendor)

$(TARGET): $(SRC) vendor gofmt
	go build -ldflags "-X main.version=$(VERSION) -X main.builddate=`date -u +%Y-%m-%dT%H:%M:%SZ`" -o $@

clean:
	$(RM) $(TARGET)

release:
	rm -rf dist/
	mkdir dist
	go get github.com/mitchellh/gox
	go get github.com/tcnksm/ghr
	CGO_ENABLED=0 gox -ldflags "-X main.version=$(VERSION) -X main.builddate=`date -u +%Y-%m-%dT%H:%M:%SZ`" -output "dist/$(TARGET)_{{.OS}}_{{.Arch}}" -os="linux"
	ghr -u erasche -replace $(VERSION) dist/

.PHONY: clean lint gofmt vet complexity qc qc_deps test clean release

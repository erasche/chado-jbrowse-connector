SRC := $(wildcard *.go)
TARGET := chado-jb-rest-api

all: $(TARGET)

deps:
	go get github.com/Masterminds/glide/...
	go install github.com/Masterminds/glide/...
	glide install

complexity: $(SRC) deps
	gocyclo -over 10 $(SRC)

vet: $(src) deps
	go vet

gofmt: $(src)
	find $(SRC) -exec gofmt -w '{}' \;

lint: $(SRC) deps
	golint $(SRC)

qc_deps:
	go get github.com/alecthomas/gometalinter
	gometalinter --install --update

qc: lint vet complexity
	#gometalinter .

test: $(SRC) deps gofmt
	go test -v $(glide novendor)

$(TARGET): $(SRC) deps gofmt
	go build -o $@

clean:
	$(RM) $(TARGET)

release:
	rm -rf dist/
	mkdir dist
	go get github.com/mitchellh/gox
	go get github.com/tcnksm/ghr
	gox -ldflags "-X main.version=`date -u +%Y-%m-%dT%H:%M:%S+00:00`" -output "dist/cjc_{{.OS}}_{{.Arch}}" -osarch="linux/amd64"

.PHONY: clean

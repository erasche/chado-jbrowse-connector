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

.PHONY: clean

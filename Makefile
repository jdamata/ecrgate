export GO111MODULE=on
VERSION=$(shell git describe --tags --candidates=1 --dirty)
BUILD_FLAGS=-ldflags="-X main.version=$(VERSION)"
SRC=$(shell find . -name '*.go')

.PHONY: all clean release install

all: ecr-gate-linux-amd64 ecr-gate-darwin-amd64

clean:
	rm -f ecr-gate ecr-gate-linux-amd64 ecr-gate-darwin-amd64

ecr-gate-linux-amd64: $(SRC)
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $@ .

ecr-gate-darwin-amd64: $(SRC)
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $@ .

install:
	rm -f acrsync
	go build $(BUILD_FLAGS) .
	mv ecr-gate ~/bin/

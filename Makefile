export GO111MODULE=on
VERSION=$(shell git describe --tags --candidates=1 --dirty)
BUILD_FLAGS=-ldflags="-X main.version=$(VERSION)"
SRC=$(shell find . -name '*.go')

.PHONY: all clean release install

all: linux darwin

clean:
	rm -f ecrgate linux darwin

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic -v ./...

linux: $(SRC)
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o ecrgate-linux .

darwin: $(SRC)
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o ecrgate-darwin .

install:
	rm -f ecrgate
	go build $(BUILD_FLAGS) .
	mv ecrgate ~/bin/

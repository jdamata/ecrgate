export GO111MODULE=on
VERSION=$(shell git describe --tags --candidates=1 --dirty)
BUILD_FLAGS=-ldflags="-X main.version=$(VERSION)"
SRC=$(shell find . -name '*.go')

.PHONY: all clean release install

all: ecrgate-linux-amd64 ecrgate-darwin-amd64

clean:
	rm -f ecrgate ecrgate-linux-amd64 ecrgate-darwin-amd64

ecrgate-linux-amd64: $(SRC)
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $@ .

ecrgate-darwin-amd64: $(SRC)
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $@ .

install:
	rm -f ecrgate
	go build $(BUILD_FLAGS) .
	mv ecrgate ~/bin/

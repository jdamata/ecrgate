FROM golang AS build-env

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /go/src/github.com/jdamata/ecrgate
ADD . /go/src/github.com/jdamata/ecrgate
RUN go build -a -tags netgo -ldflags '-w' -o /bin/ecrgate

FROM alpine
COPY --from=build-env /bin/ecrgate /ecrgate
ENTRYPOINT ["/ecrgate"]
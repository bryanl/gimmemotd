FROM golang:1.9.2

ARG version=latest
ENV VERSION=$version

WORKDIR /go/src/github.com/bryanl/gimmemotd
COPY . /go/src/github.com/bryanl/gimmemotd
RUN CGO_ENABLED=0 go build \
    -ldflags="-X main.version=${VERSION}" \
    -o gimmemotd-server \
    github.com/bryanl/gimmemotd/cmd/server

FROM alpine:3.6
RUN apk --no-cache add ca-certificates
WORKDIR /root
COPY --from=0 /go/src/github.com/bryanl/gimmemotd/gimmemotd-server .
ENTRYPOINT [ "/root/gimmemotd-server" ]
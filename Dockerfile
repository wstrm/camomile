FROM golang:1.13-alpine
LABEL authors="Philip Hjortsberg <philip@hjortsberg.me>, William Wennerström <william@willeponken.me>, Edvin Sladic <edvin@sladic.se>"

ENV PATH="${PATH}:/usr/local/bin"

RUN apk add protobuf --repository=http://dl-cdn.alpinelinux.org/alpine/edge/main

WORKDIR /usr/src/dht

COPY . /usr/src/dht
RUN go get -u github.com/golang/protobuf/protoc-gen-go
RUN go generate ./...
RUN go install ./cmd/...
CMD ["dhtnode"]

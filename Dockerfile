FROM golang:1.13-alpine
LABEL authors="Philip Hjortsberg <philip@hjortsberg.me>, William Wennerström <william@willeponken.me>, Edvin Sladic <edvin@sladic.se>"

WORKDIR /usr/src/dht

COPY . /usr/src/dht
RUN wget -O /tmp/protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v3.9.1/protoc-3.9.1-linux-x86_64.zip && unzip /tmp/protoc.zip -d /usr/local/
RUN go get -u github.com/golang/protobuf/protoc-gen-go
RUN go generate ./...
RUN go install ./cmd/...
CMD ["dhtnode"]

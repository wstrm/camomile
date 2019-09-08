LABEL maintainer="Philip Hjortsberg <Hjortsberg@github>, William Wennerström <optmzr@github>, Edvin Sladic <edvinnn@github>"
FROM golang:1.13-alpine
WORKDIR /dht
COPY go.mod go.sum ./
RUN go mod download
COPY ../. /opt/dhtnode
RUN go build -o main .
CMD ["./main"]

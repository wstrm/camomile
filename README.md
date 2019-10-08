Camomile
========

## Project status
| Build status | Test coverage | Camomile docs |
|:------------:|:-------------:|:--------------|
| [![Build Status](https://circleci.com/gh/optmzr/camomile.svg?style=svg)](https://circleci.com/gh/optmzr/camomile) | [![Code Coverage](https://codecov.io/gh/optmzr/camomile/branch/master/graph/badge.svg)](https://codecov.io/gh/optmzr/camomile) | [![Camomile Documentation](https://godoc.org/github.com/optmzr/camomile?status.svg)](https://godoc.org/github.com/optmzr/camomile) |

## Build Camomile
Download and install Protoc into `/usr/local`:
```
curl -sfL -o /tmp/protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v3.9.1/protoc-3.9.1-linux-x86_64.zip
sudo unzip /tmp/protoc.zip -d /usr/local/  # Dangerous!!!
```

Get dependencies, generate and build the binaries:
```
go get -u github.com/golang/protobuf/protoc-gen-go
go generate ./...
go build ./cmd/...
```

## Run as cluster
Build the Docker container:
```
docker build . -t dhtnode:latest
```

Start the cluster script:
```
./bin/run-cluster.sh <num> # Change <num> to the number of nodes to run.
```

Done!

## REST API
### Reference
| **Method** | **Path** | **Form Fields** | **Header**       | **Code**       | **Description**                           |
|:----------:|----------|-----------------|------------------|----------------|-------------------------------------------|
| GET        | /{key}   | N/A             | Origin: {id}     | 200 OK         | Retrieves a value by its hash key.        |
| POST       | /        | value={value}   | Location: /{key} | 202 Accepted   | Saves a value in the DHT network.         |
| DELETE     | /{key}   | N/A             | N/A              | 204 No Content | Orders the DHT network to forget a value. |

### Examples
#### Save value
```
ξ curl -iF 'value=ABC, du är mina tankar' 127.0.0.1:8080/
HTTP/1.1 202 Accepted
Location: /bde0e9f6e9d3fabd5bf6849e179f0aee485630f6d5c1c4398517cc1543fb9386
Date: Mon, 07 Oct 2019 13:42:02 GMT
Content-Length: 23
Content-Type: text/plain; charset=utf-8

ABC, du är mina tankar
```

#### Retrieve value
```
ξ curl -i 127.0.0.1:8080/bde0e9f6e9d3fabd5bf6849e179f0aee485630f6d5c1c4398517cc1543fb9386
HTTP/1.1 200 OK
Origin: 3a6b713115697a45658aac4ac5eb1714e6f985cb1826d2b5cc53562e2d490157
Date: Mon, 07 Oct 2019 13:42:44 GMT
Content-Length: 23
Content-Type: text/plain; charset=utf-8

ABC, du är mina tankar
```

#### Forget value
```
ξ curl -iX DELETE 127.0.0.1:8080/bde0e9f6e9d3fabd5bf6849e179f0aee485630f6d5c1c4398517cc1543fb9386
HTTP/1.1 204 No Content
Date: Mon, 07 Oct 2019 13:44:49 GMT
Content-Length: 0
```

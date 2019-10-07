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
| **METHOD** | **PATH**   | **FORM FIELDS** | **CODE**       | **DESCRIPTION**                           |
|:----------:|------------|-----------------|----------------|-------------------------------------------|
| GET        | /{hex key} |                 | 200 OK         | Retrieves a value by its hash key.        |
| POST       | /          | value=<value>   | 202 ACCEPTED   | Saves a value in the DHT network.         |
| DELETE     | /{hex key} |                 | 204 NO CONTENT | Orders the DHT network to forget a value. |

### Examples
#### Save value
```
curl -F 'value=ABC, du är mina tankar' 127.0.0.1:8080/
```

#### Retrieve value
```
curl 127.0.0.1:8080/bde0e9f6e9d3fabd5bf6849e179f0aee485630f6d5c1c4398517cc1543fb9386
```

#### Forget value
```
curl -X DELETE 127.0.0.1:8080/bde0e9f6e9d3fabd5bf6849e179f0aee485630f6d5c1c4398517cc1543fb9386
```

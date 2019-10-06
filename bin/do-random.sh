#!/bin/sh
randomnode="$(docker ps -q | shuf -n1)"
docker exec "$randomnode" "$@"

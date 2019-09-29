#!/bin/sh

docker exec -it "$(docker ps -q | shuf -n1)" sh

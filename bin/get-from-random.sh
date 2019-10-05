#!/bin/sh

docker exec "$(docker ps -q | shuf -n1)" dhtctl -get "$1"

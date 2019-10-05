#!/bin/sh
randomnode="$(docker ps -q | shuf -n1)"
echo "Fetching value via $randomnode:"
docker exec "$randomnode" dhtctl -get "$1"

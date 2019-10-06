#!/bin/sh
randomnode="$(docker ps -q | shuf -n1)"
echo "Putting value via $randomnode:"
docker exec "$randomnode" dhtctl -put "$1"

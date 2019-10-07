#!/bin/sh

if [ -z "$1" ]
then
    echo "Please pass the number of nodes as an argument"
    exit 1
fi

numnodes=$1
networkname="camomilenet"
networkprefix="172.22.0"
port="8118"

docker build . -t dhtnode:latest
docker network create --subnet="$networkprefix.0/16" "$networkname" >/dev/null

set -e

cleanup() {
    echo "Exiting, please wait..."
    docker ps --filter network="$networkname" -q \
        | xargs -P "$numnodes" -L 1 docker stop
    echo "Bye, have a good day! :)"
}
trap cleanup EXIT

genhash() {
    printf "%s" "$1" | sha256sum | cut -f1 -d' '
}

bootnum="1"
bootip="$networkprefix.2"
bootid=$(genhash "$bootnum")

# Start N nodes and pair them.
for num in $(seq 1 "$numnodes")
do
    nodeip="$networkprefix.$((num+1))"

    # Generate node ID with node number.
    nodeid=$(genhash "$num")

    echo "Starting node: #$num, $nodeid@$nodeip:$port"
    docker run --net "$networkname" --ip "$nodeip" -t -d \
        dhtnode "$num" "$nodeid" "$nodeip:$port" \
                "$bootid" "$bootip:$port" >/dev/null &
done

printf "Waiting for all nodes to finish starting up... "
wait  # Wait for all sub processes to finish (docker run).
echo "Done!"

# Print and follow all the logs from the nodes.
docker ps -q | xargs -L 1 -P "$numnodes" -I % \
    sh -c "docker logs --follow % | sed 's/^/[%] /g'"

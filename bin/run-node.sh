#!/bin/sh
set -euo pipefail

menum=$1
meid=$2
meaddr=$3

otherid=$4
otheraddr=$5

echo "I am: #$menum, $meid@$meaddr"
dhtnode -me "$meid@$meaddr" -other "$otherid@$otheraddr"

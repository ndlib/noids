#!/bin/bash

# Start and configure a noid server.
# This script is only intended for use in a development environment.

NOIDSDIR="$HOME/goapps/noids"
NOIDSLOG="$NOIDSDIR/dev.log"

mkdir -p "$NOIDSDIR"
touch "$NOIDSLOG"
cat /dev/null > "$NOIDSLOG"

noids -log "$NOIDSLOG" &
curl --data 'name=dev&template=.rddddd' localhost:13001/pools
tail -f "$NOIDSLOG"

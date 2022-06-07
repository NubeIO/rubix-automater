#!/bin/bash

## bash run.bash -c ../config.yaml -w false

while getopts c:w:a: flag
do
    case "${flag}" in
        c) config=${OPTARG};;
        w) wipe=${OPTARG};;
        a) add=${OPTARG};;
    esac
done

echo "config: $config";
echo "wipe: $wipe";
echo "add: $add";
if $wipe; then
    nohup go run main.go --server=true --config="$config" >/dev/null 2>&1 &
    sleep 3
    go run main.go client --wipe=true
    sudo fuser -n tcp -k 8089
fi

if $add; then
    nohup go run main.go --server=true --config="$config" >/dev/null 2>&1 &
    sleep 3
    go run main.go client --add-ping=true
    sudo fuser -n tcp -k 8089
fi

go run main.go --server=true --config="$config"


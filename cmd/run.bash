#!/bin/bash

## bash run.bash -c ../config.yaml -d true -a true

while getopts c:d:a: flag
do

    case "${flag}" in
        c) config=${OPTARG};;
        d) wipe=${OPTARG};;
        a) add=${OPTARG};;
    esac
done

echo "config: $config";
echo "wipe db: $wipe";
echo "add a example pipeline: $add";
if $wipe; then
    nohup go run main.go --server=true --config="$config" >/dev/null 2>&1 &
    echo "will restart the server after delete !!!";
    sleep 3
    go run main.go client --wipe=true
    fuser -n tcp -k 8089
fi

if $add; then
    nohup go run main.go --server=true --config="$config" >/dev/null 2>&1 &
      echo "will restart the server added a pipeline !!!";
    sleep 3
    go run main.go client --add-ping=true
    sleep 6
    fuser -n tcp -k 8089
fi

go run main.go --server=true --config="$config"


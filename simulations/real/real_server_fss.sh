#!/bin/bash

id=$1

# create results dir if doesn't already exist
mkdir -p ../results

# remove stats log and create new files
rm ../results/stats*

# build server
cd ../../cmd/grpc/server
go build

# go back to simultion directory
cd - > /dev/null

# move to root
cd ../../

# run servers
for scheme in "complexPIR" "complexVPIR"; do
  for target in "email" "algo" "and" "avg"; do
    echo "##### running server $id with $scheme scheme and $target target #####"
    # run server given the correct scheme 
    cmd/grpc/server/server -id=$id -files=31 -experiment -scheme=$scheme | tee -a /simulations/results/stats_server-${id}_${scheme}_${target}.log
    wait $!
  done
done

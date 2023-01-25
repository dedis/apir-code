#!/bin/bash

# create results dir if doesn't already exist
mkdir ../results

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
for scheme in "pointPIR" "pointVPIR"; do
  echo "##### running with $scheme scheme #####"
    # run servers
    cmd/grpc/server/server -id=$1 -files=31 -experiment -scheme=$scheme | tee -a simulations/results/stats_server-$1_$scheme.log
    wait $!
done

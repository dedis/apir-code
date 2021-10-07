#!/bin/bash
export GOGC=8000

# remove stats log and create new files
rm results/stats*

# build server
cd ../cmd/grpc/server
go build

# go back to simultion directory
cd - > /dev/null

# move to root
cd ../

# run servers
echo "##### running with $2 scheme #####"
# run server given the correct scheme 
cmd/grpc/server/server -id=$1 -files=31 -experiment -scheme=$2 | tee -a simulations/results/stats_server-0_$scheme.log
wait $!

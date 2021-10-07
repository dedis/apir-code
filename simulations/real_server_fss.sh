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
echo "##### running with FSS-based scheme #####"
# run server, db is same for complexPIR and complexVPIR
cmd/grpc/server/server -id=$1 -files=31 -experiment -scheme="complexPIR" | tee -a simulations/results/stats_server-0_$scheme.log
wait $!

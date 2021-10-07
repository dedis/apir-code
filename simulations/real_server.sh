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
for scheme in "pointPIR" "pointVPIR"; do
  echo "##### running with $scheme scheme #####"
  #for cores in {1..24}; do
    #echo "##### running with $cores cores #####"
    # run servers
    cmd/grpc/server/server -id=$1 -files=31 -experiment -scheme=$scheme | tee -a simulations/results/stats_server-0_$scheme.log
    wait $!

    #echo "##### done with $cores #####"
  #done
done

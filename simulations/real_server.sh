#!/bin/bash
export GOGC=8000

# remove stats log and create new files
rm results/stats*

# build server
cd ../cmd/grpc/server
go build -race

# go back to simultion directory
cd - > /dev/null

# move to root
cd ../

# run servers
for scheme in it dpf; do
  echo "##### running with $scheme scheme #####"
  for cores in {1..12}; do
    echo "##### running with $cores cores #####"
    # run servers
    cmd/grpc/server/server -id=$1 -files=1 -experiment -cores=$cores -scheme=$scheme | tee -a simulations/results/stats_server-0_$scheme.log
    wait $!

    echo "##### done with $cores #####"
  done
done
#!/bin/bash
export GOGC=8000

# remove stats log and create new files
rm results/stats*

# build client
cd ../cmd/grpc/client
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
    # wait for servers to setup
    sleep 2
    
    # repeat experiment 10 times
    for i in {1..20}; do
      echo "##### iteration $i running with $cores cores #####"
      cmd/grpc/client/client -id=alex.braulio@varidi.com -experiment -cores=$cores -scheme=$scheme >> simulations/results/stats_client_$scheme.log
      sleep 5
    done
    echo "##### done with $cores #####"
  done
done

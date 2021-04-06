#!/bin/bash
export GOGC=300

# remove stats log and create new files
rm stats*

# build server
cd ../cmd/grpc/server
go build -race

# go back to simultion directory
cd - > /dev/null

# build client
cd ../cmd/grpc/client
go build -race

# go back to simultion directory
cd - > /dev/null

# move to root
cd ../

# run servers
for cores in {1..8}; do
  echo "##### running with $cores cores #####"

  # run servers
  cmd/grpc/server/server -id=0 -files=1 -experiment -cores=$cores -scheme="it" >> simulations/results/stats_server-0.log & pid0=$!
  cmd/grpc/server/server -id=1 -files=1 -experiment -cores=$cores -scheme="it" >> simulations/results/stats_server-1.log & pid1=$!

  # wait for server to setup
  sleep 2
  
  # repeat experiment 10 times
  for i in {1..2}; do
    echo "##### iteration $i running with $cores cores #####"
    cmd/grpc/client/client -id=alex.braulio@varidi.com -experiment -cores=$cores -scheme="it" >> simulations/results/stats_client.log
    sleep 5
  done

  # send sigterm to servers and trigger graceful stop
  kill -TERM "$pid0"
  kill -TERM "$pid1"
  echo "sleeping to let server gracefully stops..."
  sleep 5
  echo "##### done with $cores #####"
done

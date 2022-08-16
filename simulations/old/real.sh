#!/bin/bash
export GOGC=8000

# remove stats log and create new files
rm results/stats*

# build server
cd ../cmd/grpc/server
go build

# go back to simultion directory
cd - > /dev/null

# build client
cd ../cmd/grpc/client
go build

# go back to simultion directory
cd - > /dev/null

# move to root
cd ../

# run servers
for scheme in "pointPIR" "pointVPIR"; do
  echo "##### running with $scheme #####"
  #for cores in {1..12}; do
  #echo "##### running with $cores cores #####"

  # run servers
  #cmd/grpc/server/server -id=0 -files=1 -experiment -cores=$cores -scheme=$scheme >> simulations/results/stats_server-0_$scheme.log & pid0=$!
  #cmd/grpc/server/server -id=1 -files=1 -experiment -cores=$cores -scheme=$scheme >> simulations/results/stats_server-1_$scheme.log & pid1=$!
  cmd/grpc/server/server -id=0 -files=1 -experiment -scheme=$scheme >> simulations/results/stats_server-0_$scheme.log & pid0=$!
  cmd/grpc/server/server -id=1 -files=1 -experiment -scheme=$scheme >> simulations/results/stats_server-1_$scheme.log & pid1=$!

  # wait for server to setup
  sleep 2
  
  # repeat experiment 10 times
  for i in {1..10}; do
    #echo "##### iteration $i running with $cores cores #####"
    echo "  ##### iteration $i #####"
    #cmd/grpc/client/client -id=alex.braulio@varidi.com -experiment -cores=$cores -scheme=$scheme >> simulations/results/stats_client_$scheme.log
    cmd/grpc/client/client -id=alex.braulio@varidi.com -experiment -scheme=$scheme >> simulations/results/stats_client_$scheme.log
    sleep 5
  done

  # send sigterm to servers and trigger graceful stop
  curl -s '127.0.0.1:8080' > /dev/null
  echo "sleeping to let server gracefully stops..."
  sleep 5
  #echo "##### done with $cores #####"
  #done
done

#!/bin/bash
export GOGC=8000

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
  echo "running with $cores cores"
  # run servers
  go run cmd/grpc/server/main.go -id=0 -files=1 -experiment -cores=$cores & pid0=$!
  go run cmd/grpc/server/main.go -id=1 -files=1 -experiment -cores=$cores & pid1=$!

  # wait for server to setup
  sleep 30
  
  # repeat experiment 20 times
  for i in {1..50}; do
    # run client
    go run cmd/grpc/client/main.go -id=alex.braulio@varidi.com -experiment -cores=$cores
  done

  # send sigterm to servers and trigger graceful stop
  kill -TERM "$pid0"
  kill -TERM "$pid1"
  echo "sleeping to let server gracefully stops..."
  sleep 120
  echo "done with $f files"
done

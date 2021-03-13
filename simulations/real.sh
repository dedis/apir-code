#!/bin/bash
GOGC=8000

# build server
cd ../cmd/server
go build -race

# build client
cd ../client
go build -race

# move to root
cd ../../

# initialize results file
echo "" > simulations/results/real.csv

# run servers
for f in {1..10}; do
  echo "running with $f files"
  env GOGC=$GOGC go run cmd/server/main.go -id=0 -files=$f &
  pid0=$!
  env GOGC=$GOGC go run cmd/server/main.go -id=1 -files=$f &
  pid1=$!

  # run client
  #time=$(go run cmd/client/main.go -id=alex.braulio@varidi.com | grep "Wall" | cut -d ":" -f2)
  go run cmd/client/main.go -id=alex.braulio@varidi.com

  # save value
  #echo "$f,$time" >> simulations/results/real.csv

  # send sigterm to servers and trigger graceful stop
  kill -TERM "$pid0"
  kill -TERM "$pid1"
  echo "sleeping to let server gracefully stops..."
  sleep 120
  echo "done with $f files"
done

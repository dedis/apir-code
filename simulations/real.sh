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
  env GOGC=$GOGC go run cmd/server/main.go -id=0 -files=$f 2>&1 > /dev/null &
  pid0=$!
  env GOGC=$GOGC go run cmd/server/main.go -id=1 -files=$f 2>&1 > /dev/null &
  pid1=$!

  # run client
  time = $(env GOGC=$GOGC go run cmd/client/main.go -id=alex.braulio@varidi.com | grep "Wall" | cut -d ":" -f2)

  # save value
  echo "$f,$time" >> simulations/results/real.csv

  # kill servers
  kill $pid0
  kill $pid1
done

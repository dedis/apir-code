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
  env GOGC=$GOGC go run cmd/server/main.go -id=1 -files=$f &

  # run client
  time=$(go run cmd/client/main.go -id=alex.braulio@varidi.com | grep "Wall" | cut -d ":" -f2)

  # save value
  echo "$f,$time" >> simulations/results/real.csv

  # kill servers
  kill -9 $(jobs -p) > /dev/null
  wait $!
  echo "sleeping..."
  sleep 300
  echo "done"
done

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

# run servers
env GOGC=$GOGC go run cmd/server/main.go -id=0 -files=1 2>&1 > /dev/null &
pid0=$!
env GOGC=$GOGC go run cmd/server/main.go -id=1 -files=1 2>&1 > /dev/null &
pid1=$!

# run client
env GOGC=$GOGC go run cmd/client/main.go -id=alex.braulio@varidi.com | grep "Wall"

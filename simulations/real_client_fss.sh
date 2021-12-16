#!/bin/bash
echo "REMEMBER TO UPDATE SERVERS IP ADDRESSES"

export GOGC=8000

# remove stats log and create new files
rm results/stats*

# build client
cd ../cmd/grpc/client
go build 

# go back to simultion directory
cd - > /dev/null

# move to root
cd ../

# repeat experiment 30 times
repeat=20

for scheme in "complexPIR", "complexVPIR"; do
  echo "##### running with $scheme scheme #####"

  target="email"
  # init ENV var on the servers
  curl -d "${scheme},${target}" 10.90.36.38:8000
  curl -d "${scheme},${target}" 10.90.36.39:8000

  # run experiment
  for i in {1..$repeat}; do
    echo "    ##### iteration $i with scheme $scheme"
    cmd/grpc/client/client -id=".edu" -target=$target -from-end=4 -experiment -scheme=$scheme | tee -a simulations/results/stats_client_${scheme}_${target}.log
    sleep 5
  done

  target="algo"
  # init ENV var on the servers
  curl -d "${scheme},${target}" 10.90.36.38:8000
  curl -d "${scheme},${target}" 10.90.36.39:8000
  for i in {1..$repeat}; do
    echo "    ##### iteration $i with scheme $scheme"
    cmd/grpc/client/client -id="ElGamal" -target=$target -experiment -scheme=$scheme | tee -a simulations/results/stats_client_${scheme}_${target}.log
    sleep 5
  done

  # THIS IS FOR THE AND QUERY, HARDCODED IN GO
  target="and"
  curl -d "${scheme},${target}" 10.90.36.38:8000
  curl -d "${scheme},${target}" 10.90.36.39:8000
  for i in {1..$repeat}; do
    echo "    ##### iteration $i with scheme $scheme"
    cmd/grpc/client/client -and -experiment -scheme=$scheme | tee -a simulations/results/stats_client_${scheme}_and.log
    sleep 5
  done

  # THIS IS FOR THE AVG QUERY, HARDCODED IN GO
  target="avg"
  curl -d "${scheme},${target}" 10.90.36.38:8000
  curl -d "${scheme},${target}" 10.90.36.39:8000
  for i in {1..$repeat}; do
    echo "    ##### iteration $i with scheme $scheme"
    cmd/grpc/client/client -and -avg -experiment -scheme=$scheme | tee -a simulations/results/stats_client_${scheme}_and.log
    sleep 5
  done
sleep 10
done

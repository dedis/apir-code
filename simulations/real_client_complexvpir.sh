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

scheme="complexVPIR"
echo "##### running with $scheme scheme #####"

# repeat experiment 30 times
target="email"
for i in {1..30}; do
  echo "    ##### iteration $i"
  cmd/grpc/client/client -id=".org" -target=$target -from-end=4 -experiment -scheme=$scheme | tee -a simulations/results/stats_client_$scheme_$target.log
  sleep 5
done

target="algo"
for i in {1..30}; do
  echo "    ##### iteration $i"
  cmd/grpc/client/client -id="ElGamal" -target=$target -experiment -scheme=$scheme | tee -a simulations/results/stats_client_$scheme_$target.log
  sleep 5
done

target="creation"
for i in {1..30}; do
  echo "    ##### iteration $i"
  cmd/grpc/client/client -id="2020" -target=$target -experiment -scheme=$scheme | tee -a simulations/results/stats_client_$scheme_$target.log
  sleep 5
done

# TODO: continue from here
# THIS IS FOR THE AND QUERY
target="creation"
for i in {1..30}; do
  echo "    ##### iteration $i"
  cmd/grpc/client/client -id="2020" -target=$target -experiment -scheme=$scheme | tee -a simulations/results/stats_client_$scheme_$target.log
  sleep 5
done

# terminates servers
curl 10.90.36.31:8080 > /dev/null
curl 10.90.36.33:8080 > /dev/null
sleep 10

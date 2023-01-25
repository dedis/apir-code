#!/bin/bash

# terminates servers
kill_servers () {
  curl $ip_first:8080 > /dev/null
  curl $ip_second:8080 > /dev/null
  sleep 1260
}

echo "REMEMBER TO UPDATE SERVERS IP ADDRESSES"

ip_first="10.90.38.14"
ip_second="10.90.39.3"
#
# create results dir if doesn't already exist
mkdir -p ../results

# remove stats log and create new files
rm ../results/stats*

# build client
cd ../../cmd/grpc/client
go build 

# go back to simultion directory
cd - > /dev/null

# move to root
cd ../...

for scheme in "complexPIR" "complexVPIR"; do
  echo "##### running with $scheme scheme #####"

  target="email"
  echo "    ##### running with $target target #####"
  for i in {1..30}; do
    echo "    ##### iteration $i"
    cmd/grpc/client/client -id=".edu" -target=$target -from-end=4 -experiment -scheme=$scheme | tee -a simulations/results/stats_client_${scheme}_${target}.log
    sleep 5
  done
  kill_servers

  target="algo"
  echo "    ##### running with $target target #####"
  for i in {1..30}; do
    echo "    ##### iteration $i"
    cmd/grpc/client/client -id="ElGamal" -target=$target -experiment -scheme=$scheme | tee -a simulations/results/stats_client_${scheme}_${target}.log
    sleep 5
  done
  kill_servers

  # AND QUERY, HARDCODED IN GO
  target="and"
  echo "    ##### running with $target target #####"
  for i in {1..30}; do
    echo "    ##### iteration $i"
    cmd/grpc/client/client -id=".edu" -and -experiment -scheme=$scheme | tee -a simulations/results/stats_client_${scheme}_and.log
    sleep 5
  done
  kill_servers

  # AVG QUERY, HARDCODED IN GO
  target="avg"
  echo "    ##### running with $target target #####"
  for i in {1..30}; do
    echo "    ##### iteration $i"
    cmd/grpc/client/client -id=".edu" -from-end=4 -and -avg -experiment -scheme=$scheme | tee -a simulations/results/stats_client_${scheme}_avg.log
    sleep 5
  done
  kill_servers
done


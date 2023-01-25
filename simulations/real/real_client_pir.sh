#!/bin/bash

ip_first="10.90.38.14"
ip_second="10.90.39.3"

# create results dir if doesn't already exist
mkdir ../results

# remove stats log and create new files
rm ../results/stats*

# build client
cd ../../cmd/grpc/client
go build 

# go back to simultion directory
cd - > /dev/null

# move to root
cd ../..

# run servers
for scheme in "pointPIR" "pointVPIR"; do
  echo "##### running with $scheme scheme #####"
    # repeat experiment 30 times
    for i in {1..30}; do
      echo "    ##### iteration $i"
      cmd/grpc/client/client -id=alex.braulio@varidi.com -experiment -scheme=$scheme | tee -a simulations/results/stats_client_$scheme.log
      sleep 5
    done

    # terminates servers and wait for their reboot
    curl $ip_first:8080 > /dev/null
    curl $ip_second:8080 > /dev/null
    sleep 180
done

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

scheme="$1"
echo "##### running with $scheme scheme #####"

#target="email"
#for i in {1..20}; do
  #echo "    ##### iteration $i"
  #cmd/grpc/client/client -id=".edu" -target=$target -from-end=4 -experiment -scheme=$scheme | tee -a simulations/results/stats_client_${scheme}_${target}.log
  #sleep 5
#done

#target="algo_elgamal"
#for i in {1..20}; do
  #echo "    ##### iteration $i"
  #cmd/grpc/client/client -id="ElGamal" -target=$target -experiment -scheme=$scheme | tee -a simulations/results/stats_client_${scheme}_${target}.log
  #sleep 5
#done

#target="creation"
#for i in {1..20}; do
  #echo "    ##### iteration $i"
  #cmd/grpc/client/client -id="2020" -target=$target -experiment -scheme=$scheme | tee -a simulations/results/stats_client_${scheme}_${target}.log
  #sleep 5
#done

# THIS IS FOR THE AND QUERY, HARDCODED IN GO
#for i in {1..20}; do
  #echo "    ##### iteration $i"
  #cmd/grpc/client/client -and -experiment -scheme=$scheme | tee -a simulations/results/stats_client_${scheme}_and.log
  #sleep 5
#done

## THIS IS FOR THE AND QUERY, HARDCODED IN GO
#for i in {1..20}; do
  #echo "    ##### iteration $i"
  #cmd/grpc/client/client -id=".edu" -and -experiment -scheme=$scheme | tee -a simulations/results/stats_client_${scheme}_and.log
  #sleep 5
#done

# THIS IS FOR THE AVG QUERY, HARDCODED IN GO
for i in {1..20}; do
  echo "    ##### iteration $i"
  cmd/grpc/client/client -id=".edu" -from-end=4 -and -avg -experiment -scheme=$scheme | tee -a simulations/results/stats_client_${scheme}_avg.log
  sleep 5
done

# terminates servers
curl 10.90.36.38:8080 > /dev/null
curl 10.90.36.39:8080 > /dev/null
sleep 10

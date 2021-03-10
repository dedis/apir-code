#!/bin/bash
echo "" > results.csv
for workers in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 32 64 128 256 512 1024; do
  echo "running with $workers"
  go run ./answer.go -workers=$workers >> results.csv
done
exit 0

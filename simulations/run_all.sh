#!/bin/bash

# create results dir if not already present
mkdir -p results

# run experiments
make -s run_simul config=vpirSingleVectorKB.toml > results/vpirSingleVectorKB.json
make -s run_simul config=vpirSingleVectorMB.toml > results/vpirSingleVectorMB.json
make -s run_simul config=vpirMultiVectorKB.toml > results/vpirMultiVectorKB.json
make -s run_simul config=vpirMultiVectorMB.toml > results/vpirSingleVectorMB.json

# produce plots

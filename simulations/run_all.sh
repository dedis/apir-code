#!/bin/bash

# create results dir if not already present
mkdir -p results

# run experiments
make -s run_simul config=vpirSingleVectorKB.toml > results/vpirSingleVectorKB.json
make -s run_simul config=vpirSingleVectorMB.toml > results/vpirSingleVectorMB.json

# produce plots

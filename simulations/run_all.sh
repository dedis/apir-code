#!/bin/bash

# create results dir if not already present
mkdir -p results

# run experiments
make -s run_simul config=vpirSingleVector.toml
make -s run_simul config=vpirMultiVector.toml
make -s run_simul config=vpirSingleMatrix.toml
make -s run_simul config=vpirMultiMatrix.toml

# produce plots

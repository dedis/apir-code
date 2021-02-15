#!/bin/bash

# create results dir if not already present
mkdir -p results

# run experiments
make -s run_simul config=vpirSingleVector.toml
make -s run_simul config=vpirMultiVector.toml

# produce plots

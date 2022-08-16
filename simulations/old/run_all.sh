#!/usr/bin/env bash

# create results dir if not already present
mkdir -p results

# run experiments
#make -s run_simul config=vpirSingleVector.toml
#make -s run_simul config=vpirMultiVector.toml
#make -s run_simul config=vpirMultiVectorBlock.toml
#make -s run_simul config=vpirSingleMatrix.toml
make -s run_simul config=vpirMultiMatrixBlock.toml
make -s run_simul config=vpirMultiVectorBlockDPF.toml
make -s run_simul config=pirMatrix.toml
make -s run_simul config=pirDPF.toml

# clean if we finish
#clean

# produce plots

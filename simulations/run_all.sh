#!/bin/bash

# prepend general simulation parameters to specific simulation parameters
for file in $(find . -maxdepth 1 -iname '*.toml' -not -name 'simul.toml'); do
  cat simul.toml >> $file
done

# trap CTRL-c and call ctrl_c()
trap clean INT

# horrible hack to keep specific simulation's toml files clean
function clean() {
  tmpfile=$(mktemp)
  for file in $(find . -maxdepth 1 -iname '*.toml' -not -name 'simul.toml'); do
    head -n -$(wc -l < simul.toml) $file > $tmpfile
    cat ${tmpfile} > $file
  done
  rm -f ${tmpfile}
  exit
}

# create results dir if not already present
mkdir -p results

# run experiments
#make -s run_simul config=vpirSingleVector.toml
#make -s run_simul config=vpirMultiVector.toml
make -s run_simul config=vpirMultiVectorBlockLength16.toml
#make -s run_simul config=vpirSingleMatrix.toml
make -s run_simul config=vpirMultiMatrix.toml
make -s run_simul config=vpirMultiVectorBlockLength16DPF.toml

# clean if we finish
clean

# produce plots

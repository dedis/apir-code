#!/bin/bash
INPUT=data/random_id_key.csv
OLDIFS=$IFS
IFS=','

# change directory
cd ..

# launch client
make run_client scheme=dpf

# fed client
[ ! -f $INPUT ] && { echo "$INPUT file not found"; exit 99; }
while read id key
do
  echo $id
done < $INPUT
IFS=$OLDIFS

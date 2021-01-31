#!/bin/bash
INPUT=data/random_id_key.csv
OLDIFS=$IFS
IFS=','

# change directory
cd ..

# create pipe to redirect ids later
mkfifo client_pipe
# launch client
make run_client scheme=dpf < client_pipe &

# fed client
[ ! -f $INPUT ] && { echo "$INPUT file not found"; exit 99; }
while read id key
do
  echo $id > client_pipe
done < $INPUT
IFS=$OLDIFS

# remove client pipe after experiment
rm client_pipe

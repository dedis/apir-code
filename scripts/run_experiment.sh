#!/bin/bash
on_exit() {
  # remove client pipe after experiment
  rm client_pipe
  # reset IFS
  IFS=$2
  exit
}

trap 'on_exit $client_pipe $OLDIFS' SIGTERM
trap 'on_exit $client_pipe $OLDIFS' SIGINT

INPUT=data/random_id_key.csv
OLDIFS=$IFS
IFS=','

# change directory
cd ..

# create pipe to redirect ids later
client_pipe=client_pipe
mkfifo $client_pipe

# launch client
make run_client scheme=dpf < $client_pipe &

# fed client
[ ! -f $INPUT ] && { echo "$INPUT file not found"; exit 99; }
while read id key
do
  echo $id > $client_pipe
done < $INPUT

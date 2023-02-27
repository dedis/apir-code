#!/usr/bin/env bash

BASE_FOLDER="data"
SKS_FOLDER="sks"
cd $BASE_FOLDER
mkdir -p $SKS_FOLDER && cd $_
for i in {00..30}; do 
  wget --content-disposition https://drive.switch.ch/index.php/s/jejfxB9uJRUAmqm/download\?path\=%2F\&files\=sks-0$i.pgp 
done

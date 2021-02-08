#!/usr/bin/env bash

BASE_FOLDER="data"
SKS_FOLDER="sks-original"
cd $BASE_FOLDER
mkdir -p $SKS_FOLDER && cd $_
wget -c -r -p -e robots=off -N -l1 --cut-dirs=3 -nH http://pgp.key-server.io/dump/current/
md5sum -c metadata-sks-dump.txt
#!/usr/bin/env bash

BASE_FOLDER="data"
SKS_FOLDER="sks-tmp"
cd $BASE_FOLDER
mkdir -p $SKS_FOLDER && cd $_
wget --no-check-certificate "https://1drv.ms/u/s!Anbw-PjTZVKKkBu-tGVJDIe0q_L4?e=7m83gq"
unzip sks.zip

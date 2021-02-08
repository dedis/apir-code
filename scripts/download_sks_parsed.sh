#!/usr/bin/env bash

BASE_FOLDER="data"
SKS_FOLDER="sks-tmp"
cd $BASE_FOLDER
mkdir -p $SKS_FOLDER && cd $_
wget --no-check-certificate "https://onedrive.live.com/download?cid=8A5265D3F8F8F076&resid=8A5265D3F8F8F076%212075&authkey=AA0wsCEdJaZs6X0"
unzip sks.zip

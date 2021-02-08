#!/usr/bin/python3
import json

file="vpir_vector_oneMB.json"

data = []
with open(file) as f:
    data = json.load(f)
    for db_result in data['Results']:
        print(db_result['BlockResults'])



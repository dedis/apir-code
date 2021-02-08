#!/usr/bin/python3
import json

file="vpir_vector_oneMB.json"

data = []
query = []
answer = []
reconstruct = []
with open(file) as f:
    data = json.load(f)
    for db_result in data['Results']:
        for block_result in db_result['BlockResults']:
            query.append(block_result['Query'])
            answer.append(block_result['Answer0'])
            answer.append(block_result['Answer1'])
            reconstruct.append(block_result['Reconstruct'])



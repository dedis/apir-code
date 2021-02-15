#!/usr/bin/python3
import json
import numpy as np

file="vpirSingleVector.json"

# parse results
data = []
query = []
answer = []
reconstruct = []
total = []
with open(file) as f:
    data = json.load(f)
    for dbResult in data['Results']:
        for blockResult in dbResult['Results']:
            query.append(blockResult['Query'])
            answer.append(blockResult['Answer0'])
            answer.append(blockResult['Answer1'])
            reconstruct.append(blockResult['Reconstruct'])
        total.append(dbResult['Total'])

# compute averages
queryAvg = np.mean(query)
answerAvg = np.mean(answer)
reconstructAvg = np.mean(reconstruct)
totalAvg = np.mean(total)
print(queryAvg, answerAvg,reconstructAvg,totalAvg)

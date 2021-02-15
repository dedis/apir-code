#!/usr/bin/python3
import json
import numpy as np
import matplotlib.pyplot as plt

file="vpirSingleVector.json"

# parse results
client = []
server = []
total = []
with open(file) as f:
    data = json.load(f)
    for dbResult in data['Results']:
        client.append(0)
        server.append(0)
        for blockResult in dbResult['Results']:
            client[-1] += blockResult['Query'] + blockResult['Reconstruct']
            server[-1] += (blockResult['Answer0'] + blockResult['Answer1'])/2
        total.append(dbResult['Total'])

# compute averages
clientAvg = np.mean(client)
clientStd = np.std(client)
serverAvg = np.mean(server)
serverStd = np.std(server)
totalAvg = np.mean(total)
totalStd = np.std(total)

# plot graph
labels = ['1KB']
client_means = [clientAvg]
server_means = [serverAvg]
client_std = [clientStd]
print(clientStd)
server_std = [serverStd]
width = 0.35       # the width of the bars: can also be len(x) sequence

fig, ax = plt.subplots()

ax.bar(labels, client_means, width, yerr=client_std, label='Client')
ax.bar(labels, server_means, width, yerr=server_std, bottom=client_means, label='Server')

ax.set_ylabel('CPU time [ms]')
ax.set_title('Database size')
ax.legend()

plt.show()

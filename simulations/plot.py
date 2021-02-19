#!/usr/bin/python3
import json
import numpy as np
import matplotlib.pyplot as plt
from utils import *

resultFolder = "results/"
vpirSingleVectorFile = resultFolder + "vpirSingleVector.json"
vpirSingleMatrixFile = resultFolder + "vpirSingleMatrix.json"
vpirMultiVectorFile = resultFolder + "vpirMultiVector.json"
vpirMultiMatrixFile = resultFolder + "vpirMultiMatrix.json"

statsSV = allStats(vpirSingleVectorFile)
statsSM = allStats(vpirSingleMatrixFile)
statsMV = allStats(vpirMultiVectorFile)
statsMM = allStats(vpirMultiMatrixFile)

# plot graph
labels = []
client_means = []
client_std = []
server_means =  []
server_std = []
for dbSize in statsSV:
    labels.append(dbSize)
    client_means.append(statsSV[dbSize]["client"]["mean"])
    client_std.append(statsSV[dbSize]["client"]["std"])
    server_means.append(statsSV[dbSize]["server"]["mean"])
    server_std.append(statsSV[dbSize]["server"]["std"])

width = 0.35 # the width of the bars: can also be len(x) sequence

fig, ax = plt.subplots()

ax.bar(labels, client_means, width, yerr=client_std, label='Client')
ax.bar(labels, server_means, width, yerr=server_std, bottom=client_means, label='Server')

ax.set_ylabel('CPU time [ms]')
ax.set_title('Database size [bits]')
ax.legend()

plt.show()

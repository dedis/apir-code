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

cSV, sSV, tSV, bSV = allStats(vpirSingleVectorFile)
cSM, sSM, tSM, bSM = allStats(vpirSingleMatrixFile)
cMV, sMV, tMV, bMV = allStats(vpirMultiVectorFile)
cMM, sMM, tMM, bMV = allStats(vpirMultiMatrixFile)

# plot graph
labels = ['1KB']
client_means = [clientStats['mean']]
server_means = [serverStats['mean']]
client_std = [clientStats['std']]
server_std = [serverStats['std']]
width = 0.35 # the width of the bars: can also be len(x) sequence

fig, ax = plt.subplots()

ax.bar(labels, client_means, width, yerr=client_std, label='Client')
ax.bar(labels, server_means, width, yerr=server_std, bottom=client_means, label='Server')

ax.set_ylabel('CPU time [ms]')
ax.set_title('Database size [B]')
ax.legend()

plt.show()

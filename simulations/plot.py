#!/usr/bin/python3
import json
import numpy as np
import matplotlib.pyplot as plt
from utils import *

vpirSingleVectorFileKB = "vpirSingleVector.json"
clientStats, serverStats, totalStats = allStats(vpirSingleVectorFileKB)

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
ax.set_title('Database size')
ax.legend()

plt.show()

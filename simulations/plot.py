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

cMeansSV, cStdSV, sMeansSV, sStdSV = [], [], [], []
cMeansSM, cStdSM, sMeansSM, sStdSM = [], [], [], []
cMeansMV, cStdMV, sMeansMV, sStdMV = [], [], [], []
cMeansMM, cStdMM, sMeansMM, sStdMM = [], [], [], []

for dbSize in statsSV:
    labels.append(dbSize)

    # SV
    cMeansSV.append(statsSV[dbSize]["client"]["mean"])
    cStdSV.append(statsSV[dbSize]["client"]["std"])
    sMeansSV.append(statsSV[dbSize]["server"]["mean"])
    sStdSV.append(statsSV[dbSize]["server"]["std"])

    # SM
    cMeansSM.append(statsSM[dbSize]["client"]["mean"])
    cStdSM.append(statsSM[dbSize]["client"]["std"])
    sMeansSM.append(statsSM[dbSize]["server"]["mean"])
    sStdSM.append(statsSM[dbSize]["server"]["std"])

    # MV
    cMeansMV.append(statsMV[dbSize]["client"]["mean"])
    cStdMV.append(statsMV[dbSize]["client"]["std"])
    sMeansMV.append(statsMV[dbSize]["server"]["mean"])
    sStdMV.append(statsMV[dbSize]["server"]["std"])

    # MM
    cMeansMM.append(statsMM[dbSize]["client"]["mean"])
    cStdMM.append(statsMM[dbSize]["client"]["std"])
    sMeansMM.append(statsMM[dbSize]["server"]["mean"])
    sStdMM.append(statsMM[dbSize]["server"]["std"])

width = 0.2 # the width of the bars: can also be len(x) sequence
x = np.arange(len(labels))
fig, ax = plt.subplots()

# SV
ax.bar(x - 2*width, cMeansSV, width, yerr=cStdSV, label='Client')
ax.bar(x - 2*width, sMeansSV, width, yerr=sStdSV, bottom=cMeansSV, label='Server')

# SM
ax.bar(x - width, cMeansSM, width, yerr=cStdSM, label='Client')
ax.bar(x - width, sMeansSM, width, yerr=sStdSM, bottom=cMeansSM, label='Server')

# MV
ax.bar(x, cMeansMV, width, yerr=cStdMV, label='Client')
ax.bar(x, sMeansMV, width, yerr=sStdMV, bottom=cMeansMV, label='Server')

# MM
ax.bar(x + width, cMeansMM, width, yerr=cStdMM, label='Client')
ax.bar(x + width, sMeansMM, width, yerr=sStdMM, bottom=cMeansMM, label='Server')

ax.set_ylabel('CPU time [ms]')
ax.set_title('Database size [bits]')
ax.legend()

plt.show()

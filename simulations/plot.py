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

def plotSingleMulti():
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

        # MV
        cMeansMV.append(statsMV[dbSize]["client"]["mean"])
        cStdMV.append(statsMV[dbSize]["client"]["std"])
        sMeansMV.append(statsMV[dbSize]["server"]["mean"])
        sStdMV.append(statsMV[dbSize]["server"]["std"])

    width = 0.3 # the width of the bars: can also be len(x) sequence
    x = np.arange(len(labels))
    fig, ax = plt.subplots()

    # SV
    # ax.bar(x - width*3/2, cMeansSV, width, yerr=cStdSV, label='Client SV')
    # ax.bar(x - width*3/2, sMeansSV, width, yerr=sStdSV, bottom=cMeansSV, label='Server SV')
    ax.bar(x - width/2, cMeansSV, width, yerr=cStdSV, label='Client SV', color = 'red')
    ax.bar(x - width/2, sMeansSV, width, yerr=sStdSV, bottom=cMeansSV, label='Server SV', color = 'blue')

    # MV
    # ax.bar(x - width/2, cMeansMV, width, yerr=cStdMV, label='Client MV', hatch = '//')
    # ax.bar(x - width/2, sMeansMV, width, yerr=sStdMV, bottom=cMeansMV, label='Server MV', hatch = '//')
    ax.bar(x + width/2, cMeansMV, width, yerr=cStdMV, label='Client MV', hatch = '//', color = 'red')
    ax.bar(x + width/2, sMeansMV, width, yerr=sStdMV, bottom=cMeansMV, label='Server MV', hatch = '//', color = 'blue')

    # SM
    # ax.bar(x + width/2, cMeansSM, width, yerr=cStdSM, label='Client SM')
    # ax.bar(x + width/2, sMeansSM, width, yerr=sStdSM, bottom=cMeansSM, label='Server SM')

    # MM
    # ax.bar(x + width*3/2, cMeansMM, width, yerr=cStdMM, label='Client MM', hatch = '//')
    # ax.bar(x + width*3/2, sMeansMM, width, yerr=sStdMM, bottom=cMeansMM, label='Server MM', hatch='//')

    # Totals
    #plt.errorbar(x_obs, y_obs, 0.1, fmt='.', color='black')

    ax.set_ylabel('CPU time [ms]')
    ax.set_title('Database size [bits]')
    ax.set_xticks(x)
    ax.set_xticklabels(labels)
    ax.legend()

    prepare_for_latex()
    plt.yscale('log')
    plt.savefig('figures/single_multi.png')
    #plt.show()

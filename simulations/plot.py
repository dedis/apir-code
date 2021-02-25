#!/usr/bin/env python3
import json
import numpy as np
import matplotlib.pyplot as plt
from utils import *

resultFolder = "results/"
width = 0.3

# styles
markers = ['d', 's', 'x', '.']
linestyles = ['--', '-', ':', '-.']
patterns = ['', '.', '//']


def plotVectorMatrixDPF():
    vectorFile = resultFolder + "vpirMultiVectorBlockLength16.json"
    matrixFile = resultFolder + "vpirMultiMatrix.json"
    dpfFile = resultFolder + "vpirMultiVectorBlockLength16DPF.json"

    statsVector = allStats(vectorFile)
    statsMatrix = allStats(matrixFile)
    statsDPF = allStats(dpfFile)

    cMeanVector, cStdVector, sMeanVector, sStdVector = [], [], [], []
    cMeanMatrix, cStdMatrix, sMeanMatrix, sStdMatrix = [], [], [], []
    cMeanDPF, cStdDPF, sMeanDPF, sStdDPF = [], [], [], []
    labels = []

    for dbSize in statsDPF:
        labels.append(dbSize)

        # vector
        cMeanVector.append(statsVector[dbSize]["client"]["mean"])
        cStdVector.append(statsVector[dbSize]["client"]["std"])
        sMeanVector.append(statsVector[dbSize]["server"]["mean"])
        sStdVector.append(statsVector[dbSize]["server"]["std"])

        # matrix
        cMeanMatrix.append(statsMatrix[dbSize]["client"]["mean"])
        cStdMatrix.append(statsMatrix[dbSize]["client"]["std"])
        sMeanMatrix.append(statsMatrix[dbSize]["server"]["mean"])
        sStdMatrix.append(statsMatrix[dbSize]["server"]["std"])

        # dpf
        cMeanDPF.append(statsDPF[dbSize]["client"]["mean"])
        cStdDPF.append(statsDPF[dbSize]["client"]["std"])
        sMeanDPF.append(statsDPF[dbSize]["server"]["mean"])
        sStdDPF.append(statsDPF[dbSize]["server"]["std"])

    x = np.arange(len(labels))
    fig, ax = plt.subplots()

    # vector
    ax.bar(x - width, cMeanVector, width, yerr=cStdVector, label='Client vector')
    ax.bar(x - width, sMeanVector, width, yerr=sStdVector, bottom=cMeanVector, label='Server vector')

    # matrix
    ax.bar(x, cMeanMatrix, width, yerr=cStdMatrix, label='Client matrix')
    ax.bar(x, sMeanMatrix, width, yerr=sStdMatrix, bottom=cMeanMatrix, label='Server matrix')

    # dpf
    ax.bar(x + width, cMeanDPF, width, yerr=cStdDPF, label='Client DPF')
    ax.bar(x + width, sMeanDPF, width, yerr=sStdDPF, bottom=cMeanDPF, label='Server DPF')

    ax.set_ylabel('CPU time [ms]')
    ax.set_title('Database size [bits]')
    ax.set_xticks(x)
    ax.set_xticklabels(labels)
    ax.legend()

    prepare_for_latex()
    plt.yscale('log')
    plt.savefig('figures/vector_matrix_dpf.png')
    #plt.show()


def plotSingleMulti():
    vpirSingleVectorFile = resultFolder + "vpirSingleVector.json"
    vpirMultiVectorFile = resultFolder + "vpirMultiVector.json"

    statsSV = allStats(vpirSingleVectorFile)
    statsMV = allStats(vpirMultiVectorFile)

    # plot graph
    labels = []

    cMeansSV, cStdSV, sMeansSV, sStdSV = [], [], [], []
    cMeansMV, cStdMV, sMeansMV, sStdMV = [], [], [], []

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

    x = np.arange(len(labels))
    fig, ax = plt.subplots()

    # SV
    ax.bar(x - width/2, cMeansSV, width, yerr=cStdSV, label='Client SV', color = 'red')
    ax.bar(x - width/2, sMeansSV, width, yerr=sStdSV, bottom=cMeansSV, label='Server SV', color = 'blue')

    # MV
    ax.bar(x + width/2, cMeansMV, width, yerr=cStdMV, label='Client MV', hatch = '//', color = 'red')
    ax.bar(x + width/2, sMeansMV, width, yerr=sStdMV, bottom=cMeansMV, label='Server MV', hatch = '//', color = 'blue')

    ax.set_ylabel('CPU time [ms]')
    ax.set_title('Database size [bits]')
    ax.set_xticks(x)
    ax.set_xticklabels(labels)
    ax.legend()

    prepare_for_latex()
    plt.yscale('log')
    plt.savefig('figures/single_multi.png')
    #plt.show()


def plotVpirBenchmarksBarBw():
    schemes = ["vpirSingleVector.json", "vpirMultiVector.json", "vpirMultiVectorBlock.json"]
    labels = ["Single-bit", "Multi-bit", "Multi-bit Block"]

    Xs = np.arange(len(schemes))
    width = 0.35
    Ys, Yerr = [], []
    for scheme in schemes:
        stats = allStats(resultFolder + scheme)
        largestDbSize = sorted(stats.keys())[-1]
        Ys.append(stats[largestDbSize]['client']['cpu']['mean'] + stats[largestDbSize]['server']['cpu']['mean'])
        Yerr.append(stats[largestDbSize]['client']['cpu']['std'] + stats[largestDbSize]['server']['cpu']['std'])

    plt.style.use('grayscale')
    fig, ax1 = plt.subplots()
    color = 'black'
    ax1.set_ylabel("CPU time [ms]", color=color)
    ax1.tick_params(axis='y', labelcolor=color)
    ax1.set_xticks(Xs + width / 2)
    ax1.set_xticklabels(labels)
    ax1.bar(Xs, Ys, width, label="CPU", color=color, yerr=Yerr)
    plt.yscale('log')
    ax1.legend(fontsize=12)

    Ys, Yerr = [], []
    for scheme in schemes:
        stats = allStats(resultFolder + scheme)
        largestDbSize = sorted(stats.keys())[-1]
        Ys.append(stats[largestDbSize]['client']['bw']['mean']/1000 + stats[largestDbSize]['server']['bw']['mean']/1000)
        Yerr.append(stats[largestDbSize]['client']['bw']['std']/1000 + stats[largestDbSize]['server']['bw']['std']/1000)

    color = 'grey'
    ax2 = ax1.twinx()  # instantiate a second axes that shares the same x-axis
    ax2.set_ylabel("Bandwidth [KB]")
    ax2.bar(Xs+width, Ys, width, label="Bandwidth", color=color, yerr=Yerr)
    ax2.legend(loc=5, fontsize=12)

    # fig.tight_layout()  # otherwise the right y-label is slightly clipped
    plt.yscale('log')
    plt.title("Retrieval of 256B of data from 125KB DB")
    plt.savefig('cpu_bw.eps', format='eps', dpi=300)
    # plt.show()


def plotVpirBenchmarks():
    schemes = ["vpirSingleVector.json", "vpirMultiVector.json", "vpirMultiVectorBlock.json"]
    labels = ["Single-bit", "Multi-bit", "Multi-bit Block"]
    colors = ['black', 'lightgrey', 'grey']

    fig, ax = plt.subplots()
    plt.style.use('grayscale')

    width = 0.15
    dbSizes = sorted([int(size/8000) for size in allStats(resultFolder + schemes[0]).keys()])
    Xs = np.arange(len(dbSizes))
    bars = [[]]*len(schemes)
    for i, scheme in enumerate(schemes):
        stats = allStats(resultFolder + scheme)
        for j, dbSize in enumerate(sorted(stats.keys())):
            Y = stats[dbSize]['client']['cpu']['mean'] + stats[dbSize]['server']['cpu']['mean']
            Yerr = stats[dbSize]['client']['cpu']['std'] + stats[dbSize]['server']['cpu']['std']
            bars[i] = ax.bar(j+i*width, Y, width, color=colors[i], yerr=Yerr)
            ax.annotate(f'{Y:.1f}',
                        xy=(j+i*width, Y),
                        xytext=(0, 5),  # 3 points vertical offset
                        textcoords="offset points",
                        ha='center', va='bottom')

    ax.set_ylabel("CPU time [ms]")
    ax.set_xlabel("DB size [KB]")
    ax.set_xticks(Xs + width*(len(schemes)-1)/2)
    ax.set_xticklabels(dbSizes)
    ax.legend(bars, labels, fontsize=12)
    ax.spines['top'].set_visible(False)
    ax.spines['right'].set_visible(False)
    ax.spines['left'].set_visible(False)

    plt.tick_params(left=False, labelleft=False)
    plt.yscale('log')
    plt.title("Retrieval of 256B data from a DB of different sizes")
    # plt.savefig('benchmarks.eps', format='eps', dpi=300, transparent=True)
    plt.show()


def plotVpirPerformance():
    colors = ['darkred', 'darkorange', 'darkgreen', 'darkblue']
    devcolors = ['mistyrose', 'papayawhip', 'honeydew', 'ghostwhite']
    schemes = ["vpirMultiMatrix.json", "vpirMultiVectorBlockDPF.json", "pirMatrix.json", "pirDPF.json"]
    labels = ["VPIR rebalanced", "VPIR DPF", "PIR rebalanced", "PIR DPF"]

    fig, ax = plt.subplots()
    i = 0
    for scheme in schemes:
        stats = allStats(resultFolder + scheme)
        Xs, Ys, Yerr = [], [], []
        for dbSize in sorted(stats.keys()):
            Xs.append(dbSize/8000000)
            Ys.append(stats[dbSize]['client']['cpu']['mean'] + stats[dbSize]['server']['cpu']['mean'])
            Yerr.append(stats[dbSize]['client']['cpu']['std'] + stats[dbSize]['server']['cpu']['std'])

        print(Ys)
        ax.errorbar(Xs, Ys, yerr=Yerr, color=colors[i], label=labels[i], marker=markers[i], linestyle=linestyles[i])
        # ax.fill_between(Xs, Yerrdown, Yerrup, facecolor=devcolors[i])
        i += 1

    ax.legend(fontsize=12)
    ax.set_ylabel('CPU time [ms]')
    ax.set_xlabel('Database size [MB]')
    ax.spines['top'].set_visible(False)
    ax.spines['right'].set_visible(False)
    # plt.xscale('log')
    plt.axis()
    # plt.savefig('performance.eps', format='eps', dpi=300, transparent=True)
    plt.show()


if __name__ == "__main__":
    plotVpirBenchmarks()
    # plotVpirPerformance()

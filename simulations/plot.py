#!/usr/bin/env python3
import json
import numpy as np
import matplotlib.pyplot as plt
from utils import *

resultFolder = "results/"
width = 0.3

# colors and constants
colors = ['#E2DC27', '#071784', '#077C0F', '#BC220A']
devcolors = ['#FFFDCD', '#CDE1FF', '#D4FFE3', '#FFDFD1']
markers = ['d', 's', 'x', '.']
linestyles = ['--', ':', '-', '-.']
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


def plotVpirBenchmarksLinear():
    schemes = ["vpirSingleVector.json", "vpirMultiVector.json", "vpirMultiVectorBlockLength16.json"]
    labels = ["Single-bit", "Multi-bit", "Multi-bit Blocks"]

    i = 0
    for scheme in schemes:
        stats = allStats(resultFolder + scheme)
        Xs, Ys, Yerrup, Yerrdown = [], [], [], []
        for dbSize in sorted(stats.keys()):
            Xs.append(dbSize)
            Ys.append(stats[dbSize]['client']['cpu']['mean'] + stats[dbSize]['server']['cpu']['mean'])
            std = stats[dbSize]['client']['cpu']['std'] + stats[dbSize]['server']['cpu']['std']
            Yerrup.append(Ys[-1] + std)
            Yerrdown.append(Ys[-1] - std)

        print(Xs)
        print(Ys)
        plt.loglog(Xs, Ys, color=colors[i], label=labels[i], marker=markers[i], linestyle=linestyles[i])
        plt.fill_between(Xs, Yerrdown, Yerrup, facecolor=devcolors[i])
        i += 1

    plt.legend(loc='upper left', fontsize=12)
    plt.ylabel('CPU time [ms]')
    plt.xlabel('Database size [bits]')
    # plt.xscale('log')
    plt.axis()
    plt.savefig('boring_lines.eps', format='eps', dpi=300)
    # plt.show()


def plotVpirBenchmarksBarBw():
    schemes = ["vpirSingleVector.json", "vpirMultiVector.json", "vpirMultiVectorBlockLength16.json"]
    labels = ["Single-bit", "Multi-bit", "Multi-bit Blocks"]

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
    plt.title("Retrieval of 2KB of data from 200KB DB")
    plt.savefig('cpu_bw.eps', format='eps', dpi=300)
    # plt.show()


def plotVpirBenchmarksBar():
    schemes = ["vpirSingleVector.json", "vpirMultiVector.json", "vpirMultiVectorBlockLength16.json"]
    labels = ["Single-bit", "Multi-bit", "Multi-bit Blocks"]
    colors = ['lightgrey', 'darkgrey', 'dimgrey', 'black']

    fig, ax = plt.subplots()
    plt.style.use('grayscale')

    Xs = np.arange(len(schemes))
    width = 0.15
    Ys, Yerr = [], []
    dbSizes = ["10KB", "50KB", "100KB", "200KB"]
    bars = [[]]*len(dbSizes)
    for i, scheme in enumerate(schemes):
        stats = allStats(resultFolder + scheme)
        for j, dbSize in enumerate(sorted(stats.keys())):
            Ys = stats[dbSize]['client']['cpu']['mean'] + stats[dbSize]['server']['cpu']['mean']
            Yerr = stats[dbSize]['client']['cpu']['std'] + stats[dbSize]['server']['cpu']['std']
            bars[j] = ax.bar(i+j*width, Ys, width, color=colors[j], yerr=Yerr)

    ax.set_ylabel("CPU time [ms]")
    ax.set_xticks(Xs + width*4 / 3)
    ax.set_xticklabels(labels)
    plt.yscale('log')
    ax.legend(bars, dbSizes, fontsize=12)

    # fig.tight_layout()  # otherwise the right y-label is slightly clipped
    # plt.yscale('log')
    plt.title("Retrieval of 2KB of data from DBs of different size")
    plt.savefig('multi_cpu.eps', format='eps', dpi=300)
    # plt.show()


if __name__ == "__main__":
    # plotSingleMulti()
    # plotVectorMatrixDPF()
    # plotVpirBenchmarksLinear()
    plotVpirBenchmarksBar()

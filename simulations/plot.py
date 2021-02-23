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


def plotVpirBenchmarks():
    schemes = ["vpirSingleVector.json", "vpirMultiVector.json", "vpirMultiVectorBlockLength16.json"]
    labels = ["Single-bit", "Multi-bit", "Multi-bit Blocks"]

    i = 0
    for scheme in schemes:
        stats = allStats(resultFolder + scheme)
        Xs, Ys, Yerrup, Yerrdown = [], [], [], []
        for dbSize in sorted(stats.keys()):
            Xs.append(dbSize)
            Ys.append(stats[dbSize]["client"]["mean"] + stats[dbSize]["server"]["mean"])
            std = stats[dbSize]["client"]["std"] + stats[dbSize]["server"]["std"]
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
    plt.axis()
    # plt.savefig('figures/multi_benchmarks.eps', format='eps', dpi=300)
    plt.show()


if __name__ == "__main__":
    # plotSingleMulti()
    # plotVectorMatrixDPF()
    plotVpirBenchmarks()



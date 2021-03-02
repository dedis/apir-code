#!/usr/bin/env python3
import argparse
import matplotlib.pyplot as plt
from utils import *

resultFolder = "results/"
width = 0.3

# styles
markers = ['.', 'x', 'd', 's']
linestyles = ['-', '--', ':', '-.']
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
    labels = ["Single-bit", "Multi-bit", "Block"]
    colors = ['black', 'grey', 'lightgrey']

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
    # ax.spines['left'].set_visible(False)

    # plt.tick_params(left=False, labelleft=False)
    plt.tight_layout()
    plt.yscale('log')
    # plt.title("Retrieval of 256B data from a DB of different sizes")
    plt.savefig('multi_benchmarks.eps', format='eps', dpi=300, transparent=True)
    # plt.show()


def plotVpirPerformance():
    colors = ['darkred', 'darkblue', 'darkorange', 'darkgreen']
    devcolors = ['mistyrose', 'ghostwhite', 'papayawhip', 'honeydew']
    schemes = ["vpirMultiMatrixBlock.json", "vpirMultiVectorBlockDPF.json", "pirMatrix.json", "pirDPF.json"]
    labels = ["CPU rebalanced", "BW rebalanced", "CPU DPF", "BW DPF"]

    fig, ax1 = plt.subplots()
    ax1.set_ylabel('VPIR/PIR CPU ratio')
    ax1.set_xlabel('Database size [MB]')
    ax1.spines['top'].set_visible(False)
    ax1.spines['right'].set_visible(False)

    ax2 = ax1.twinx()  # instantiate a second axes that shares the same x-axis
    ax2.set_ylabel("VPIR/PIR bandwidth ratio")
    ax2.spines['top'].set_visible(False)
    ax2.spines['right'].set_linestyle((0, (5, 10)))
    # ax2.set_yscale('log')

    # Save PIR values first so we can divide by them later
    Xpir, Ypir, Ypirbw = [], [], []
    for scheme in schemes[int(len(schemes)/2):]:
        stats = allStats(resultFolder + scheme)
        for dbSize in sorted(stats.keys()):
            Xpir.append(dbSize/8000000)
            Ypir.append(stats[dbSize]['client']['cpu']['mean'] + stats[dbSize]['server']['cpu']['mean'])
            Ypirbw.append(stats[dbSize]['client']['bw']['mean'] + stats[dbSize]['server']['bw']['mean'])

    j = 0
    for i, scheme in enumerate(schemes[:int(len(schemes)/2)]):
        stats = allStats(resultFolder + scheme)
        Xs, Ys, Ybw = [], [], []
        for dbSize in sorted(stats.keys()):
            if Xpir[j] == dbSize/8000000:
                Xs.append(dbSize/8000000)
                Ys.append((stats[dbSize]['client']['cpu']['mean'] + stats[dbSize]['server']['cpu']['mean']) / Ypir[j])
                Ybw.append((stats[dbSize]['client']['bw']['mean'] + stats[dbSize]['server']['bw']['mean']) / Ypirbw[j])
            else:
                print("Xs do not align")
                break
            j += 1

        ax1.plot(Xs, Ys, color=colors[i], marker=markers[i], linestyle=linestyles[0], label=labels[2*i])
        ax2.plot(Xs, Ybw, color=colors[i], marker=markers[i], linestyle=linestyles[1], label=labels[2*i+1])

    handles, labels = [(a + b) for a, b in zip(ax1.get_legend_handles_labels(), ax2.get_legend_handles_labels())]
    plt.legend(handles, labels, bbox_to_anchor=(0.95, 0.7), loc='center right',
               ncol=1, borderaxespad=0.)
    plt.tight_layout()
    # plt.title("CPU and bandwidth VPIR-to-PIR ratio")
    plt.savefig('multi_performance.eps', format='eps', dpi=300, transparent=True)
    # plt.show()


# -----------Argument Parser-------------
parser = argparse.ArgumentParser()
parser.add_argument("-e", "--expr", type=str, help="experiment to plot: benchmarks, performance", required=True)

args = parser.parse_args()
EXPR = args.expr

if __name__ == "__main__":
    prepare_for_latex()
    if EXPR == "benchmarks":
        plotVpirBenchmarks()
    elif EXPR == "performance":
        plotVpirPerformance()
    else:
        print("Unknown experiment: choose between benchmarks and performance")

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
        Ys.append(
            stats[largestDbSize]['client']['bw']['mean'] / 1000 + stats[largestDbSize]['server']['bw']['mean'] / 1000)
        Yerr.append(
            stats[largestDbSize]['client']['bw']['std'] / 1000 + stats[largestDbSize]['server']['bw']['std'] / 1000)

    color = 'grey'
    ax2 = ax1.twinx()  # instantiate a second axes that shares the same x-axis
    ax2.set_ylabel("Bandwidth [KB]")
    ax2.bar(Xs + width, Ys, width, label="Bandwidth", color=color, yerr=Yerr)
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
    dbSizes = sorted([int(size / 8000) for size in allStats(resultFolder + schemes[0]).keys()])
    Xs = np.arange(len(dbSizes))
    bars = [[]] * len(schemes)
    for i, scheme in enumerate(schemes):
        stats = allStats(resultFolder + scheme)
        for j, dbSize in enumerate(sorted(stats.keys())):
            Y = stats[dbSize]['client']['cpu']['mean'] + stats[dbSize]['server']['cpu']['mean']
            Yerr = stats[dbSize]['client']['cpu']['std'] + stats[dbSize]['server']['cpu']['std']
            bars[i] = ax.bar(j + i * width, Y, width, color=colors[i], yerr=Yerr)
            ax.annotate(f'{Y:.1f}',
                        xy=(j + i * width, Y),
                        xytext=(0, 5),  # 3 points vertical offset
                        textcoords="offset points",
                        ha='center', va='bottom')

    ax.set_ylabel("CPU time [ms]")
    ax.set_xlabel("DB size [KB]")
    ax.set_xticks(Xs + width * (len(schemes) - 1) / 2)
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
    GB = 1e9
    MB = 1e6
    colors = ['darkred', 'darkblue', 'darkorange', 'darkgreen']
    schemes = ["vpirMultiMatrixBlock.json", "merkleMatrix.json", "pirMatrix.json", "vpirMultiVectorBlockDPF.json",
               "merkleDPF.json", "pirDPF.json"]
    labels = ["VPIR matrix", "Merkle matrix", "PIR matrix", "VPIR DPF", "Merkle DPF", "PIR DPF"]
    # schemes = ["vpirMultiMatrixBlock.json", "pirMatrix.json", "vpirMultiVectorBlockDPF.json", "pirDPF.json",]
    # labels = ["VPIR matrix", "PIR matrix", "VPIR DPF", "PIR DPF"]

    fig, ax = plt.subplots()
    ax.set_ylabel('Requests/second')
    ax.set_xlabel('Requests/GB')
    ax.spines['top'].set_visible(False)
    ax.spines['right'].set_visible(False)

    # ax2.spines['right'].set_linestyle((0, (5, 10)))
    ax.set_xscale('log')
    ax.set_yscale('log')
    ax.yaxis.grid(True)

    table = defaultdict(list)

    for i, scheme in enumerate(schemes):
        Xs, Ys = [], []
        stats = allStats(resultFolder + scheme)
        for dbSize in sorted(stats.keys()):
            bw = stats[dbSize]['client']['bw']['mean'] + stats[dbSize]['server']['bw']['mean']
            cpu = stats[dbSize]['client']['cpu']['mean'] + stats[dbSize]['server']['cpu']['mean']
            table[dbSize].append((cpu, bw/1000))
            # print("%.2f & %d & " % (1000/cpu, GB/bw), end="")
            Xs.append(GB/bw)
            Ys.append(1000/cpu)
            ax.annotate(str(int(int(dbSize)/(8*MB)))+"MB", xy=(GB/bw, 1000/cpu), xytext=(-20, 5), color=colors[int(i/(len(schemes)/2))], textcoords='offset points')

        # print(Xs)
        # print(Ys)
        ax.plot(Xs, Ys, color=colors[int(i/(len(schemes)/2))], marker=markers[0], linestyle=linestyles[i%int(len(schemes)/2)], label=labels[i])

    for size, values in table.items():
        print(str(int(int(size)/(8*MB)))+"\\,MB", end=" ")
        for value in values:
            if value[0] > 5 and value[1] > 5:
                print("& %d & %d " % (round(value[0]), round(value[1])), end="")
            elif value[0] < 5 and value[1] > 5:
                print("& %.2f & %d " % (value[0], value[1]), end="")
            elif value[0] > 5 and value[1] < 5:
                print("& %d & %.2f " % (value[0], value[1]), end="")
            else:
                print("& %.2f & %.2f " % (value[0], value[1]), end="")
        print("\\\\")

    ax.legend(loc='lower center')
    # ax.legend(bbox_to_anchor=(0, 1.02, 1, 0.2), loc="lower left",
    #           mode="expand", borderaxespad=0, ncol=4)
    plt.tight_layout()
    # plt.savefig('multi_performance.eps', format='eps', dpi=300, transparent=True)
    plt.show()


# def plotVpirPerformance():
#     colors = ['darkred', 'darkblue', 'darkorange', 'darkgreen']
#     devcolors = ['mistyrose', 'ghostwhite', 'papayawhip', 'honeydew']
#     schemes = ["vpirMultiMatrixBlock.json", "vpirMultiVectorBlockDPF.json", "pirMatrix.json", "pirDPF.json"]
#     labels = ["CPU rebalanced", "BW rebalanced", "CPU DPF", "BW DPF"]
#
#     fig, ax1 = plt.subplots()
#     ax1.set_ylabel('VPIR/PIR CPU ratio')
#     ax1.set_xlabel('Database size [MB]')
#     ax1.spines['top'].set_visible(False)
#     ax1.spines['right'].set_visible(False)
#
#     ax2 = ax1.twinx()  # instantiate a second axes that shares the same x-axis
#     ax2.set_ylabel("VPIR/PIR bandwidth ratio")
#     ax2.spines['top'].set_visible(False)
#     ax2.spines['right'].set_linestyle((0, (5, 10)))
#     # ax2.set_yscale('log')
#
#     # Save PIR values first so we can divide by them later
#     Xpir, Ypir, Ypirbw = [], [], []
#     for scheme in schemes[int(len(schemes)/2):]:
#         stats = allStats(resultFolder + scheme)
#         for dbSize in sorted(stats.keys()):
#             Xpir.append(dbSize/8000000)
#             Ypir.append(stats[dbSize]['client']['cpu']['mean'] + stats[dbSize]['server']['cpu']['mean'])
#             Ypirbw.append(stats[dbSize]['client']['bw']['mean'] + stats[dbSize]['server']['bw']['mean'])
#
#     j = 0
#     for i, scheme in enumerate(schemes[:int(len(schemes)/2)]):
#         stats = allStats(resultFolder + scheme)
#         Xs, Ys, Ybw = [], [], []
#         for dbSize in sorted(stats.keys()):
#             if Xpir[j] == dbSize/8000000:
#                 Xs.append(dbSize/8000000)
#                 Ys.append((stats[dbSize]['client']['cpu']['mean'] + stats[dbSize]['server']['cpu']['mean']) / Ypir[j])
#                 Ybw.append((stats[dbSize]['client']['bw']['mean'] + stats[dbSize]['server']['bw']['mean']) / Ypirbw[j])
#             else:
#                 print("Xs do not align")
#                 break
#             j += 1
#
#         ax1.plot(Xs, Ys, color=colors[i], marker=markers[i], linestyle=linestyles[0], label=labels[2*i])
#         ax2.plot(Xs, Ybw, color=colors[i], marker=markers[i], linestyle=linestyles[1], label=labels[2*i+1])
#
#     handles, labels = [(a + b) for a, b in zip(ax1.get_legend_handles_labels(), ax2.get_legend_handles_labels())]
#     plt.legend(handles, labels, bbox_to_anchor=(0.95, 0.7), loc='center right',
#                ncol=1, borderaxespad=0.)
#     plt.tight_layout()
#     # plt.title("CPU and bandwidth VPIR-to-PIR ratio")
#     plt.savefig('multi_performance.eps', format='eps', dpi=300, transparent=True)
#     # plt.show()


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

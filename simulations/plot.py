#!/usr/bin/env python3
import argparse

import matplotlib.lines as mlines
import matplotlib.patches as mpatches
import matplotlib.pyplot as plt

from utils import *

resultFolder = "final_results/"

# styles
markers = ['.', '*', 'd', 's']
linestyles = ['-', '--', ':', '-.']
patterns = ['', '//', '.']

GB = pow(1024, 3)
MB = pow(1024, 2)
KB = 1024
LatticeRotKeysLen = 39322025
LatticeCiphertextLen = 393221

def plotVpirBenchmarks():
    # schemes = ["vpirSingleVector.json", "vpirMultiVector.json", "vpirMultiVectorBlock.json"]
    # labels = ["Single-bit (§4.2)", "Single-element (§4.4)", "Block (§4.4)"]
    # colors = ['black', 'grey', 'lightgrey']

    schemes = ["vpirSingleVector.json", "vpirMultiVectorBlock.json"]
    labels = ["Single-bit (§4.2)", "Block (§4.4)"]
    colors = ['black', 'darkgrey']

    fig, ax = plt.subplots()
    plt.style.use('grayscale')

    width = 0.2
    dbSizes = sorted([int(size / (8 * KB)) for size in allStats(resultFolder + schemes[0]).keys()])
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
                        xytext=(0, 5),  # 5 points vertical offset
                        textcoords="offset points",
                        ha='center', va='bottom')

    ax.set_ylabel("CPU time [ms]")
    ax.set_xlabel("DB size [KiB]")
    ax.set_xticks(Xs + width * (len(schemes) - 1) / 2)
    ax.set_xticklabels(dbSizes)
    ax.legend(bars, labels)
    ax.spines['top'].set_visible(False)
    ax.spines['right'].set_visible(False)

    # plt.tick_params(left=False, labelleft=False)
    plt.tight_layout()
    plt.yscale('log')
    plt.savefig('multi_benchmarks.eps', format='eps', dpi=300, transparent=True)
    # plt.show()


def plotVpirPerformanceBars():
    colors = ['dimgray', 'darkgray', 'lightgrey']
    schemes = ["pirMatrix.json", "merkleMatrix.json", "vpirMultiMatrixBlock.json",
               "pirDPF.json", "merkleDPF.json", "vpirMultiVectorBlockDPF.json"]
    schemeLabels = ["PIR", "Merkle", "VPIR"]
    optimizationLabels = ["Matrix", "DPF"]

    fig, ax = plt.subplots()
    # ax.set_ylabel('Ratio to PIR Matrix latency')
    ax.set_ylabel('Ratio to PIR Matrix bandwidth')
    ax.set_xlabel('Database size [MB]')
    ax.spines['top'].set_visible(False)
    ax.spines['right'].set_visible(False)

    ax.set_yscale('log')
    # ax.yaxis.grid(True)

    width = 0.15
    dbSizes = sorted([int(size / 8000000) for size in allStats(resultFolder + schemes[0]).keys()])
    Xs = np.arange(len(dbSizes))
    bars = [[]] * len(schemes)

    # each db size is normalized by PIR matrix latency of that db size
    baselines = defaultdict(int)
    for i, scheme in enumerate(schemes):
        stats = allStats(resultFolder + scheme)
        for j, dbSize in enumerate(sorted(stats.keys())):
            # Y = stats[dbSize]['client']['cpu']['mean'] + stats[dbSize]['server']['cpu']['mean']
            Y = stats[dbSize]['client']['bw']['mean'] + stats[dbSize]['server']['bw']['mean']
            if i == 0:
                baselines[dbSize] = Y
            # Yerr = stats[dbSize]['client']['cpu']['std'] + stats[dbSize]['server']['cpu']['std']
            # bars[i] = ax.bar(j + i * width, Y, width, yerr=Yerr, color=colors[i % 3], hatch=patterns[int(i / 3)])
            # if i != 0:
            bars[i] = ax.bar(j + i * width, Y / baselines[dbSize], width, color=colors[i % 3],
                             hatch=patterns[int(i / 3)])
            # else:
            #     bars[i] = ax.bar(j + i * width, Y / baselines[dbSize], width, color='darkred',
            #                      hatch=patterns[int(i / 3)])
            ax.annotate(rounder(Y / baselines[dbSize]),
                        xy=(j + i * width, Y / baselines[dbSize]),
                        xytext=(0, 0),  # 5 points vertical offset
                        rotation=45,
                        textcoords="offset points",
                        ha='center', va='bottom')

    ax.set_xticks(Xs + width * (len(schemes) - 1) / 2)
    ax.set_xticklabels(dbSizes)
    ax.plot([0, len(Xs)], [1, 1], color='black')

    handles = []
    for i, label in enumerate(schemeLabels):
        handles.append(mpatches.Patch(color=colors[i], label=label))
    for i, label in enumerate(optimizationLabels):
        handles.append(mpatches.Patch(facecolor='white', edgecolor='black', hatch=patterns[i], label=label))

    # ax.legend(handles=handles, loc='upper left', ncol=2)
    # ax.legend(handles=handles, bbox_to_anchor=(0.01, 1.2, 0.39, 0.1), loc="upper left",
    ax.legend(handles=handles, bbox_to_anchor=(0.01, 1.08, 0.94, 0.1), loc="upper left",
              mode="expand", borderaxespad=0, ncol=5)
    plt.tight_layout()
    # plt.savefig('multi_performance_bar_cpu.eps', format='eps', dpi=300, transparent=True)
    plt.savefig('multi_performance_bar_bw.eps', format='eps', dpi=300, transparent=True)
    # plt.show()


def plotVpirPerformanceLines():
    colors = ['darkred', 'darkgreen', 'darkblue', 'darkorange']
    schemes = ["pirMatrix.json", "merkleMatrix.json", "vpirMultiMatrixBlock.json",
               "pirDPF.json", "merkleDPF.json", "vpirMultiVectorBlockDPF.json"]
    # schemes = ["merkleMatrix.json", "vpirMultiMatrixBlock.json",
    #            "merkleDPF.json", "vpirMultiVectorBlockDPF.json"]

    fig, ax = plt.subplots()
    ax.set_xlabel('Requests/s')
    ax.set_ylabel('Requests/GB')
    ax.spines['top'].set_visible(False)
    ax.spines['right'].set_visible(False)

    # ax2.spines['right'].set_linestyle((0, (5, 10)))
    ax.set_xscale('log')
    ax.set_yscale('log')
    ax.yaxis.grid(True)
    ax.xaxis.grid(True)
    # ax.invert_xaxis()
    # ax.invert_yaxis()

    cpuTable = defaultdict(list)
    bwTable = defaultdict(list)
    dbSizes = [str(int(int(size) / (8 * MB))) + "MB" for size in allStats(resultFolder + schemes[0]).keys()]

    for i, scheme in enumerate(schemes):
        Xs, Ys = [], []
        stats = allStats(resultFolder + scheme)
        for j, dbSize in enumerate(sorted(stats.keys())):
            bw = stats[dbSize]['client']['bw']['mean'] + stats[dbSize]['server']['bw']['mean']
            cpu = stats[dbSize]['client']['cpu']['mean'] + stats[dbSize]['server']['cpu']['mean']
            cpuTable[dbSize].append(cpu)
            bwTable[dbSize].append(bw / 1000)
            Xs.append(1000 / cpu)
            Ys.append(GB / bw)
            annotation = str(int(int(dbSize) / (8 * MB))) + "MB"
            if int(dbSize) == 8 * GB:
                annotation = "1GB"
            ax.annotate(annotation, xy=(1000 / cpu, GB / bw), xytext=(-20, 5),
                        color=colors[i % int(len(schemes) / 2)], textcoords='offset points')
            ax.plot(Xs[-1], Ys[-1], color=colors[i % int(len(schemes) / 2)], marker=".")
            # ax.annotate(str(int(int(dbSize) / (8 * MB))) + "MB", xy=(1000/cpu, GB/bw), xytext=(-20, 5),
            #             color=colors[i % int(len(schemes) / 2)], textcoords='offset points')
            # ax.plot(Xs[-1], Ys[-1], color=colors[i % int(len(schemes) / 2)], marker=markers[j])

        ax.plot(Xs, Ys, color=colors[i % int(len(schemes) / 2)],
                linestyle=linestyles[int(i / (len(schemes) / 2))])

    print_latex_table_separate(cpuTable, int(len(schemes) / 2), get_size_in_mib)
    print("")
    print_latex_table_separate(bwTable, int(len(schemes) / 2), get_size_in_mib)

    schemeLabels = ["No integrity", "Merkle", "VPIR"]
    optimizationLabels = ["Matrix", "DPF"]
    handles = []
    for i, label in enumerate(schemeLabels):
        handles.append(mpatches.Patch(color=colors[i], label=label))
    for i, label in enumerate(optimizationLabels):
        handles.append(mlines.Line2D([], [], color='black',
                                     linestyle=linestyles[i], label=label))
    # for i, size in enumerate(dbSizes):
    #     handles.append(mlines.Line2D([], [], color='black',
    #                                  marker=markers[i], label=size))
    ax.legend(handles=handles, loc='center left')

    # ax.legend(handles=handles, bbox_to_anchor=(0, 1.08, 1, 0.2), loc="lower left",
    #           mode="expand", borderaxespad=0, fancybox=True, ncol=3)
    plt.tight_layout()
    # plt.savefig('multi_performance.eps', format='eps', dpi=300, transparent=True)
    # plt.show()


def plotSingle():
    schemes = ["computationalPir.json", "computationalVpir.json"]
    labels = ["None", "Atomic"]
    cpuTable = defaultdict(list)
    bwTable = defaultdict(list)
    for i, scheme in enumerate(schemes):
        stats = allStats(resultFolder + scheme)
        for j, dbSize in enumerate(sorted(stats.keys())):
            bw = stats[dbSize]['client']['bw']['mean'] + stats[dbSize]['server']['bw']['mean']
            if scheme == schemes[0]:
                bw -= LatticeRotKeysLen
            cpu = stats[dbSize]['client']['cpu']['mean'] + stats[dbSize]['server']['cpu']['mean']
            cpuTable[dbSize].append(cpu / 1000)
            bwTable[dbSize].append(bw/MB)

    print_latex_table_separate(cpuTable, len(schemes), get_size_in_bits)
    print("")
    print_latex_table_separate(bwTable, len(schemes), get_size_in_bits)


def plotReal():
    schemes = ["it", "dpf", "merkle-it", "merkle-dpf", "pir-it", "pir-dpf"]
    labels = ["Atomic", "Merkle", "PIR"]
    dbSizes = [12.485642 ,11.650396 ,11.907099,11.122669,11.702634 ,10.918602]

    fig, ax = plt.subplots()

    for i, scheme in enumerate(schemes):
        logServers = [
                "stats_server-0_" + scheme + ".log", 
                "stats_server-1_" + scheme + ".log"]

        statsServers = []
        for l in logServers:
            statsServers.append(parseLog(resultFolder + l))

        # combine answers bandwidth
        answers = dict()
        for core in statsServers[0]:
                answers[core] = [sum(x) for x in zip(statsServers[0][core]["answer"], statsServers[1][core]["answer"])]

        # get client stats
        statsClient = parseLog(resultFolder + "stats_client_" + scheme + ".log")

        queries = dict()
        latencies = dict()
        for c in statsClient:
            queries[c] = statsClient[c]["queries"]
            latencies[c] = statsClient[c]["latency"]

        
        # take medians
        #answersMean = meanFromDict(answers)
        #queriesMean = meanFromDict(queries)
        ping = 0.375815 # ms
        latencyMean = meanFromDict(latencies)
        #bestLatency = latencyMean[24] + ping
        worstLatency = latencyMean[1] + ping
        
        if i % 2 == 0:
            #print(labels[int(i/2)], "&",  rounder2(worstLatency), "&", rounder(bestLatency), "&", end=" ")
            print(labels[int(i/2)], "&",  round(worstLatency, 2), "&", round(dbSizes[i], 2), "&", end=" ")
        else:
            #print(rounder2(worstLatency), "&", rounder2(bestLatency), "\\\\") 
            print(round(worstLatency, 2), "&", round(dbSizes[i], 2), "\\\\") 


def print_latex_table_separate(results, numApproaches, get_printable_size):
    for size, values in results.items():
        print(get_printable_size(size), end=" ")
        for i, value in enumerate(values):
            print("& %s " % rounder2(value), end="")
            # we need to compute the overhead over the baseline that is always at position i%numApproaches==0
            if i % numApproaches != 0:
                print("& %s$\\times$ " % rounder2(value / values[int(i / numApproaches) * numApproaches]), end="")
        print("\\\\")


def print_latex_table_joint(results, numApproaches):
    for size, values in results.items():
        print(str(int(int(size) / (8 * MB))) + "\\,MB", end=" ")
        for i, value in enumerate(values):
            print("& %s & %s " % (rounder2(value[0]), rounder2(value[1])), end="")
            # compute overhead
            if i % numApproaches == numApproaches - 1:
                print("& %s & %s " % (rounder2(value[0] / values[i - 1][0]), rounder2(value[1] / values[i - 1][1])),
                      end="")
        print("\\\\")


def get_size_in_mib(bits):
    return str(int(int(bits) / (8 * MB))) + "\\,MiB"


def get_size_in_bits(bits):
    return str(int(bits / 1e6)) + "\\,M"


def rounder(x):
    if x > 3:
        return f'{x:.0f}'
    # elif x > 1:
    #     return f'{x:.1f}'
    else:
        return f'{x:.1f}'


def rounder2(x):
    if x > 5:
        return f'{round(x):,.0f}'
    else:
        return f'{round(x, 1):,.1f}'


# def plotVpirPerformanceRatio():
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
parser.add_argument(
    "-e",
    "--expr",
    type=str,
    help="experiment to plot: benchmarks, performance, single, real",
    required=True)

args = parser.parse_args()
EXPR = args.expr

if __name__ == "__main__":
    prepare_for_latex()
    if EXPR == "benchmarks":
        plotVpirBenchmarks()
    elif EXPR == "performance":
        plotVpirPerformanceLines()
        # plotVpirPerformanceBars()
    elif EXPR == "single":
        plotSingle()
    elif EXPR == "real":
        plotReal()
    else:
        print("Unknown experiment: choose between benchmarks and performance")

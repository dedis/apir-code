#!/usr/bin/env python3
import argparse

import matplotlib.lines as mlines
import matplotlib.patches as mpatches
import matplotlib.pyplot as plt

from utils import *

resultFolder = "final_results/"
#resultFolder = "results/"

print("plotting from", resultFolder)

# styles
markers = ['.', '*', 'd', 's']
linestyles = ['-', '--', ':', '-.']
patterns = ['', '//', '.']

GB = pow(1024, 3)
bitsToGB = 0.000000000125
MB = pow(1024, 2)
KB = 1024
GiB = 8589935000
MiB = 1048576
LatticeRotKeysLen = 39322025
LatticeCiphertextLen = 393221

def plotPoint():
    schemes = ["pirClassic.json", "pirMerkle.json"]
    schemeLabels = ["Unauthenticated", "Authenticated"]

    fig, axs = plt.subplots(2, sharex=True)

    cpuArray = []
    bwArray = []
    x = []
    for i, scheme in enumerate(schemes):
        stats = allStats(resultFolder + scheme) 
        cpuArray.append([])
        bwArray.append([])
        for j, dbSize in enumerate(sorted(stats.keys())):
            # means
            cpuMean = stats[dbSize]['client']['cpu']['mean'] + stats[dbSize]['server']['cpu']['mean']
            bwMean = stats[dbSize]['client']['bw']['mean'] + stats[dbSize]['server']['bw']['mean']
            cpuArray[i].append(cpuMean/1000)
            bwArray[i].append(bwMean/MB)


        axs[0].plot(
                [x/GiB for x in sorted(stats.keys())], 
                cpuArray[i], 
                color='black', 
                marker=markers[int(i / (len(schemes) / 2))],
                linestyle=linestyles[int(i / (len(schemes) / 2))],
                label=schemeLabels[int(i / (len(schemes) / 2))]
        )
        axs[1].plot(
                [x/GiB for x in sorted(stats.keys())], 
                bwArray[i], 
                color='black', 
                marker=markers[int(i / (len(schemes) / 2))],
                linestyle=linestyles[int(i / (len(schemes) / 2))],
                label=schemeLabels[int(i / (len(schemes) / 2))]
        )

    ratioCPU = [cpuArray[1][i]/cpuArray[0][i] for i in range(len(cpuArray[0]))]
    ratioBW = [bwArray[1][i]/bwArray[0][i] for i in range(len(cpuArray[0]))]

    print("ratios CPU:", max(ratioCPU))
    print("ratios BW:", max(ratioBW))

    # cosmetics
    axs[0].set_ylabel('CPU time [s]')
    axs[0].set_xticks([int(x/GiB) for x in sorted(stats.keys())]), 
    axs[1].set_ylabel('Bandwidth [MiB]')
    axs[1].set_xlabel('DB size [GiB]')
    axs[0].legend(bbox_to_anchor=(0., 1.02, 1., .102), loc='lower left',
           ncol=2, mode="expand", borderaxespad=0.)

    plt.tight_layout(h_pad=1.5)
    plt.savefig('figures/point.eps', format='eps', dpi=300, transparent=True)

def plotComplex(): 
    schemes = ["fss.json", "authfss.json"]
    schemeLabels = ["Unauthenticated", "Authenticated"]

    fig, axs = plt.subplots(2, sharex=True)

    cpuArray = []
    bwArray = []
    x = []
    for i, scheme in enumerate(schemes):
        stats = allStats(resultFolder + scheme) 
        cpuArray.append([])
        bwArray.append([])
        for j, dbSize in enumerate(sorted(stats.keys())):
            # means
            cpuMean = stats[dbSize]['client']['cpu']['mean'] + stats[dbSize]['server']['cpu']['mean']
            bwMean = stats[dbSize]['client']['bw']['mean'] + stats[dbSize]['server']['bw']['mean']
            cpuArray[i].append(cpuMean/1000)
            bwArray[i].append(bwMean/1024)

        axs[0].plot(
                [x for x in sorted(stats.keys())], 
                cpuArray[i], 
                color='black', 
                linestyle=linestyles[int(i / (len(schemes) / 2))],
                label=schemeLabels[int(i / (len(schemes) / 2))]
        )

        axs[1].plot(
                [x for x in sorted(stats.keys())], 
                bwArray[i], 
                color='black', 
                linestyle=linestyles[int(i / (len(schemes) / 2))],
                label=schemeLabels[int(i / (len(schemes) / 2))]
        )

    print("mean ratio CPU:", np.median([cpuArray[1][i]/cpuArray[0][i] for i in range(len(cpuArray[0]))]))
    print("mean ratio BW:", np.median([bwArray[1][i]/bwArray[0][i] for i in range(len(cpuArray[0]))]))

    # cosmetics
    axs[0].set_ylabel('CPU time [s]')
    axs[0].set_xticks([x for x in sorted(stats.keys())])
    axs[1].set_ylabel('Bandwidth [KiB]')
    axs[1].set_xlabel('Function-secret-sharing input size [bytes]')
    axs[0].legend(bbox_to_anchor=(0., 1.02, 1., .102), loc='lower left',
           ncol=2, mode="expand", borderaxespad=0.)

    plt.tight_layout()
    #plt.show()
    plt.savefig('figures/complex.eps', format='eps', dpi=300, transparent=True)

def plotComplexBars(): 
    schemes = ["fss.json", "authfss.json"]
    schemeLabels = ["Unauthenticated", "Authenticated"]

    width = 0.35  # the width of the bars
    fig, ax = plt.subplots()

    cpuArray = []
    bwArray = []
    for i, scheme in enumerate(schemes):
        stats = allStats(resultFolder + scheme) 
        cpuArray.append([])
        bwArray.append([])
        for j, dbSize in enumerate(sorted(stats.keys())):
            # means
            cpuMean = stats[dbSize]['client']['cpu']['mean'] + stats[dbSize]['server']['cpu']['mean']
            bwMean = stats[dbSize]['client']['bw']['mean'] + stats[dbSize]['server']['bw']['mean']
            cpuArray[i].append(cpuMean/1000)
            bwArray[i].append(bwMean/1024)

    print("mean ratio CPU:", np.median([cpuArray[1][i]/cpuArray[0][i] for i in range(len(cpuArray[0]))]))
    print("mean ratio BW:", np.median([bwArray[1][i]/bwArray[0][i] for i in range(len(cpuArray[0]))]))
    
    ratioCPU = [cpuArray[1][i]/cpuArray[0][i] for i in range(len(cpuArray[0]))]
    ratioBW = [bwArray[1][i]/bwArray[0][i] for i in range(len(cpuArray[0]))]

    x = np.arange(len(ratioCPU))
    
    rects1 = ax.bar(
            x - width/2, 
            ratioCPU, width, 
            label='CPU', 
            color='0.3', 
            edgecolor='black', 
            )
    rects2 = ax.bar(
            x + width/2, 
            ratioBW, width, 
            label='Bandwidth',
            color='0.7', 
            edgecolor = 'black',
            )
    ax.axhline(y = 1, color ='black', linestyle = '--')

    # cosmetics
    ax.set_ylabel('Relative overhead between \n authenticated and unauthenticated PIR')
    ax.set_xticks(x, [x for x in sorted(stats.keys())])
    ax.set_xlabel('Function-secret-sharing input size [bytes]')
    ax.set_ylim(bottom=0.9)
    ax.legend(bbox_to_anchor=(0., 1.02, 1., .102), loc='lower left',
           ncol=2, mode="expand", borderaxespad=0.)

    plt.tight_layout()
    #plt.show()
    plt.savefig('figures/complex_bars.eps', format='eps', dpi=300, transparent=True)

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


def plotRealComplex():
    schemes = [
        "complexPIR_email", 
        "complexVPIR_email", 
        "complexPIR_algo", 
        "complexVPIR_algo", 
        "complexPIR_and", 
        "complexVPIR_and", 
        "complexPIR_avg", 
        "complexVPIR_avg", 
    ]

    core = -1 # only a single core
    c = core

    fig, ax = plt.subplots()

    bwUnauth = 0
    for i, scheme in enumerate(schemes):
        logServers = [
                "stats_server-0_" + scheme + ".log", 
                "stats_server-1_" + scheme + ".log"]

        statsServers = []
        for l in logServers:
            statsServers.append(parseLog(resultFolder + l))

        # combine answers bandwidth
        answers = dict()
        answers[core] = [sum(x) for x in zip(statsServers[0][core]["answer"], statsServers[1][core]["answer"])]
        serversBW = answers[core]

        # get client stats
        statsClient = parseLog(resultFolder + "stats_client_" + scheme + ".log")

        queries = dict()
        latencies = dict()
        queries[c] = statsClient[c]["queries"]
        latencies[c] = statsClient[c]["latency"]

        clientBW = queries[c]
        userTime = latencies[c]
        
        totalBW = [sum(x) for x in zip(serversBW, clientBW)]
        
        t = round(np.median(userTime), 2)
        bw = rounder2(np.median(totalBW)/1000)
        if i % 2 == 0:
            if scheme ==  "complexPIR_email":
                print('\\texttt{COUNT} of emails ending with ".edu" & &', t, "&", end=" ") 
            elif scheme == "complexPIR_algo":
                print('\\texttt{COUNT} of ElGamal keys & &', t, "&", end=" ") 
            elif scheme == "complexPIR_and":
                print('\\texttt{COUNT} of keys created in 2019 AND ending with ".edu" & &', t, "&", end=" ") 
            elif scheme == "complexPIR_avg":
                print('\\texttt{AVG} lifetime of keys for emails ending with ".edu" & &', t, "&", end=" ") 
            else: 
                print("unknow scheme")

            bwUnauth = bw
        else:
            print(t, "&", bwUnauth, "&", bw, "\\\\")
        
def plotReal():
    schemes = ["merkle-dpf", "pir-dpf"]
    labels = ["Authenticated", "Unauthenticated"]
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
            print(labels[int(i/2)], "&",  round(worstLatency, 2), "&", end=" ")
        else:
            #print(rounder2(worstLatency), "&", rounder2(bestLatency), "\\\\") 
            print(round(worstLatency, 2), "\\\\") 


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
    if EXPR == "point":
        plotPoint()
    elif EXPR == "complex":
        plotComplex()
    elif EXPR == "complexBars":
        plotComplexBars()
    elif EXPR == "single":
        plotSingle()
    elif EXPR == "real":
        plotReal()
    elif EXPR == "realcomplex":
        plotRealComplex()
    else:
        print("Unknown experiment: choose between benchmarks and performance")

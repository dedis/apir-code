#!/usr/bin/env python3
import argparse

import matplotlib.lines as mlines
import matplotlib.patches as mpatches
import matplotlib.pyplot as plt

from utils import *

resultFolder = "final_results/"
resultFolder = "results/"

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

def cpuMean(stats, key):
    # always plotted in seconds
    return (stats[key]['client']['cpu']['mean'] + stats[key]['server']['cpu']['mean'])/1000

def bwMean(stats, key):
    return stats[key]['client']['bw']['mean'] \
            + stats[key]['server']['bw']['mean']

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
            # store means
            cpuArray[i].append(cpuMean(stats, dbSize))
            bwArray[i].append(bwMean(stats, dbSize)/MB)


        axs[0].plot(
                [x/GiB for x in sorted(stats.keys())], 
                cpuArray[i], 
                color='black', 
                marker=markers[int(i / (len(schemes) / 2))],
                linestyle=linestyles[int(i / (len(schemes) / 2))],
                label=schemeLabels[int(i / (len(schemes) / 2))],
                linewidth=0.5,
        )
        axs[1].plot(
                [x/GiB for x in sorted(stats.keys())], 
                bwArray[i], 
                color='black', 
                marker=markers[int(i / (len(schemes) / 2))],
                linestyle=linestyles[int(i / (len(schemes) / 2))],
                label=schemeLabels[int(i / (len(schemes) / 2))],
                linewidth=0.5,
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
            # store means
            cpuArray[i].append(cpuMean(stats, dbSize))
            bwArray[i].append(bwMean(stats, dbSize)/1024)

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
            # store means 
            cpuArray[i].append(cpuMean(stats, dbSize))
            bwArray[i].append(bwMean(stats, dbSize)/1024)

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

def plotSingle():
    size_to_unit = {8192: "1 KiB", 8389000: "1 MiB", 8590000000: "1 GiB"}
    schemes = ["computationalDH.json", "computationalLWE.json"]
    labels = ["Auth. DH", "Auth. LWE"]
    cpuTable = defaultdict(list)
    bwTable = defaultdict(list)
    for i, scheme in enumerate(schemes):
        stats = allStats(resultFolder + scheme)
        for j, dbSize in enumerate(sorted(stats.keys())):
            bw = bwMean(stats, dbSize) 
            cpu = cpuMean(stats, dbSize)
            cpuTable[dbSize].append(cpu*1000) # print in seconds
            bwTable[dbSize].append(bw*9.765625e-4)
    
    for size, values in cpuTable.items():
        # DH is not compute for 1GiB
        if size == 8590000000:
            print(size_to_unit[size], end = " ")
            print("& x & N/A & x$\\times$ &", end = " ")
            print(rounder2(values[0]), " & x$\\times$", end = " ")
            # bw
            print("& x & N/A & x$\\times$ &", end = " ")
            print(rounder2(bwTable[size][0]), " & x$\\times$ \\\\")
        else:
            print(size_to_unit[size], end = " ")
            print("& x & ", rounder2(values[0]), " & x$\\times$ &", end = " ")
            print(rounder2(values[1]), " & x$\\times$", end = " ")
            # bw
            print("& x & ", rounder2(bwTable[size][0]), " & x$\\times$ &", end = " ")
            print(rounder2(bwTable[size][1]), " & x$\\times$ \\\\")

    # print_latex_table_joint(cpuTable, len(schemes), get_size_in_bits)
    # print("")
    # print_latex_table_joint(bwTable, len(schemes), get_size_in_bits)


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
            print(labels[int(i/2)], "&",  round(worstLatency, 2), "&", end=" ")
        else:
            print(round(worstLatency, 2), "\\\\") 


def plotMulti():
    schemes = ["pirClassicMulti.json", "pirMerkleMulti.json"]
    schemeLabels = ["Unauthenticated", "Authenticated"]

    fig, axs = plt.subplots(2, sharex=True) 

    cpuArray = []
    bwArray = []
    x = []
    for i, scheme in enumerate(schemes):
        stats = allStats(resultFolder + scheme)
        cpuArray.append([])
        bwArray.append([])
        for j, numServers in enumerate(sorted(stats.keys())):
            # means
            cpuArray[i].append(cpuMean(stats, numServers))
            bwArray[i].append(bwMean(stats, numServers)/MB)

        axs[0].plot(
                sorted(stats.keys()), 
                cpuArray[i], 
                color='black', 
                marker=markers[int(i / (len(schemes) / 2))],
                linestyle=linestyles[int(i / (len(schemes) / 2))],
                label=schemeLabels[int(i / (len(schemes) / 2))],
                linewidth=0.5,
        )
        axs[1].plot(
                sorted(stats.keys()), 
                bwArray[i], 
                color='black', 
                marker=markers[int(i / (len(schemes) / 2))],
                linestyle=linestyles[int(i / (len(schemes) / 2))],
                label=schemeLabels[int(i / (len(schemes) / 2))],
                linewidth=0.5,
        )

    # cosmetics
    axs[0].set_ylabel('CPU time [s]')
    axs[0].set_xticks(sorted(stats.keys())), 
    axs[1].set_ylabel('Bandwidth [MiB]')
    axs[1].set_xlabel('Number of servers')
    axs[0].legend(bbox_to_anchor=(0., 1.02, 1., .102), loc='lower left',
           ncol=2, mode="expand", borderaxespad=0.)

    plt.tight_layout(h_pad=1.5)
    plt.savefig('figures/multi.eps', format='eps', dpi=300, transparent=True)


def plotPreprocessing():
    scheme = "preprocessingMerkle.json"
    
    results = dict()
    # parse results, these are different from the others 
    # since we only store preprocessing time
    # using the server's Answer time
    with open(resultFolder + scheme) as f:
        data = json.load(f)
        # dbResults is a dict containing all the results for the
        # given size of the db, expressed in bits
        dbResults = data['Results']
        # iterate the (size, results) pairs
        for dbSize, dbResult in dbResults.items():
            results[int(dbSize)] = []
            for r in dbResult:
                results[int(dbSize)].append(r['CPU'][0]['Answers'][0])

    # take mean
    for k, v in results.items():
        results[k] = np.mean(v)

    plt.plot(
    range(len(results)), 
    [x/1000 for x in sorted(results.values())],
    color='black', 
    marker=markers[0],
    linestyle=linestyles[0],
    linewidth=0.5,
    )

    # fig, axs = plt.subplots() 

    # cosmetics
    plt.xticks(range(len(results)), [int(x/GiB) for x in sorted(results.keys())])
    plt.ylabel('CPU time [s]')
    # axs.set_xticks(sorted(results.keys())), 
    plt.xlabel('Database size [GiB]')
    # axs.legend(bbox_to_anchor=(0., 1.02, 1., .102), loc='lower left',
           # ncol=2, mode="expand", borderaxespad=0.)

    plt.tight_layout(h_pad=1.5)
    plt.savefig('figures/preprocessing.eps', format='eps', dpi=300, transparent=True)

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
    elif EXPR == "multi":
        plotMulti()
    elif EXPR == "preprocessing":
        plotPreprocessing()
    else:
        print("Unknown experiment: choose between benchmarks and performance")

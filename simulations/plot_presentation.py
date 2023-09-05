#!/usr/bin/env python3
import argparse
import os

import matplotlib
import matplotlib.font_manager as font_manager
import matplotlib.pyplot as plt

from utils import *

resultFolder = "final_results/"
#resultFolder = "results/"

print("plotting from", resultFolder)

# styles
markers = ['.', '*', 'd', 's']
linestyles = ['-', '--', ':', '-.']
patterns = ['', '//', '.']

MB = pow(1024, 2)
GiB = 1 << 33

def cpuMean(stats, key):
    # always plotted in seconds
    return (stats[key]['client']['cpu']['mean'] + stats[key]['server']['cpu']['mean'])/1000

def bwMean(stats, key):
    return stats[key]['client']['bw']['mean'] \
            + stats[key]['server']['bw']['mean']


def sci_notation(number, sig_fig=2):
    ret_string = "{0:.{1:d}e}".format(number, sig_fig)
    a, b = ret_string.split("e")
    # remove leading "+" and strip leading zeros
    b = int(b)
    return "$" + a + " \\cdot 10^{" + str(b) + "}$ "
 
def plotSingle():
    size_to_unit = {1<<13: "1KiB", 1<<23: "1MiB", 1<<33: "1GiB"}
    base_latex = "\\multirow{3}{*}"
    size_to_latex = {
            1 << 13: base_latex + "{1 KiB}",
            1 << 23: base_latex + "{1 MiB}",
            1 << 33: base_latex + "{1 GiB}",
        }
    size_to_units_latex = {
            1 << 13: ["[KiB]", "[KiB]", "[ms]"],
            1 << 23: ["[MiB]", "[KiB]", "[ms]"],
            1 << 33: ["[MiB]", "[KiB]", "[s]"],
        }
    size_to_multipliers = {
            1 << 13: [1.0, 1.0, 1.0],
            1 << 23: [1/1024.0, 1.0, 1.0],
            1 << 33: [1/1024.0, 1.0, 1/1000.0],
        }
    schemes = ["computationalLWE128.json", "computationalLWE.json", "simplePIR.json"]
    names = ['LWE', 'LWE$^+$', 'SimplePIR [HHCMV]']
    cpuTable = {}
    bwTable = {}
    digestTable = {}
    for i, scheme in enumerate(schemes):
        if scheme == "spiral":
            continue
        stats = allStats(resultFolder + scheme)
        cpuTable[scheme] = {}
        bwTable[scheme] = {}
        digestTable[scheme] = {}
        for j, dbSize in enumerate(sorted(stats.keys())):
            bw = bwMean(stats, dbSize) 
            cpu = cpuMean(stats, dbSize)
            cpuTable[scheme][dbSize] = cpu*1000*1000 # store in ms (already divided by 1000 in function)
            bwTable[scheme][dbSize] = (bw/1024.0) # KiB, since everything is already in bytes
            digestTable[scheme][dbSize] = stats[dbSize]['digest']/1024.0 # KiB, since everything is already in bytes

    cpuData = [list(cpuTable[schemes[i]].values()) for i in range(len((schemes)))]

    offData = [list(digestTable[schemes[i]].values()) for i in range(len(schemes))]

    onData = [list(bwTable[schemes[i]].values()) for i in range(len(schemes))]

    # plot numerical result
    fig, (ax1, ax2, ax3) = plt.subplots(1, 3)
    width = 0.2
    x = np.arange(len(list(size_to_unit.values())))
     
    # Set the default color cycle
    plt.rcParams['axes.prop_cycle'] = mpl.cycler(color=["r", "k", "c"])
    for i in range(len(schemes)):
        ax1.bar(x + i*width, cpuData[i], width, label=names[i]) # , color='#000080', label='Case-1', yerr=data_std[:,0])
        ax2.bar(x + i*width, offData[i], width, label=names[i]) # , color='#000080', label='Case-1', yerr=data_std[:,0])
        ax3.bar(x + i*width, onData[i], width, label=names[i]) # , color='#000080', label='Case-1', yerr=data_std[:,0])

    ax2.bar(x + 3*width, [1, 1024, 1048576], width, label='DB')

    axs = [ax1, ax2, ax3]
    for ax in axs:
        #ax.set_xticks(x + width + width/2, fontsize=11)
        ax.set_xticks(x + width + width/2)
        ax.set_xticklabels(list(size_to_unit.values()))
        ax.set_xlabel('Database size')
        # Axis label slightly bigger
        ax.xaxis.label.set_size(14)
        ax.set_yscale('log')
    ax1.set_ylabel('User time [ms]')
    ax1.yaxis.label.set_size(14)
    ax1.yaxis.set_tick_params(labelsize=13)
    ax2.set_ylabel('Offline bandwidth [KiB]')
    ax2.yaxis.label.set_size(14)
    ax2.yaxis.set_tick_params(labelsize=13)
    ax3.set_ylabel('Online bandwidth [KiB]')
    ax3.yaxis.label.set_size(14)
    ax3.yaxis.set_tick_params(labelsize=13)

    # legend
    #plt.legend(loc="upper left")
    fig.legend(names, bbox_to_anchor=(0.14, 1, 0.8, .102), loc='lower left',
           ncol=len(names), mode="expand", borderaxespad=0., fontsize=13)

    # for item in ([ax.xaxis.label, ax.yaxis.label] +
    #              ax.get_xticklabels() + ax.get_yticklabels()):
    #     item.set_fontsize(20)
    plt.tight_layout(h_pad=1.5)
    plt.savefig('single_bar_presentation.png', format='png', dpi=300,bbox_inches="tight")

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

    c = -1 # only a single core

    fig, ax = plt.subplots()

    bwUnauth = 0
    tUnauth = 0
    for i, scheme in enumerate(schemes):
        logServers = [
                "stats_server-0_" + scheme + ".log", 
                "stats_server-1_" + scheme + ".log"]

        statsServers = []
        for l in logServers:
            statsServers.append(parseLog(resultFolder + l))

        # combine answers bandwidth
        answers = dict()
        answers[c] = [sum(x) for x in zip(statsServers[0][c]["answer"], statsServers[1][c]["answer"])]
        serversBW = answers[c]

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
            tUnauth = t
        else:
            print(t, "&", round(float(t)/float(tUnauth), 2), "&", bwUnauth, "&", bw, "&", round(float(bw)/float(bwUnauth), 2), "\\\\")
        
def plotReal():
    schemes = ["pointVPIR", "pointPIR"]
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
        ping = 0.375815 # ms
        latencyMean = meanFromDict(latencies)
        latency = latencyMean[-1] + ping
       
        print("wall-clock time needed to retrieve a PGP public-key:")
        print(labels[i], ":", round(latency, 2), "sec")

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
        results[k] = np.median(v)

    plt.plot(
        range(len(results)), 
        [x/1000 for x in sorted(results.values())],
        color='black', 
        marker=markers[0],
        linestyle=linestyles[0],
        linewidth=0.5,
    )

    # cosmetics
    plt.xticks(range(len(results)), [int(x/GiB) for x in sorted(results.keys())])
    plt.ylabel('CPU time [s]')
    plt.xlabel('Database size [GiB]')
    plt.tight_layout(h_pad=1.5)
    plt.savefig('figures/preprocessing.eps', format='eps', dpi=300, transparent=True)
    print("figure preprocessing.eps succesfully saved in figures/")

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
    # prepare_for_latex()
    matplotlib.rcParams['text.usetex'] = True
    matplotlib.rcParams['font.family'] = 'serif'
    matplotlib.rcParams['font.serif'] = 'Times'
    if not os.path.exists("figures"):
        os.makedirs("figures")
    if not os.path.exists("results"):
        os.makedirs("results")

    if EXPR == "single":
        plotSingle()
        #plotSingleRatios()
    elif EXPR == "real":
        plotReal()
    elif EXPR == "realcomplex":
        plotRealComplex()
    elif EXPR == "multi":
        plotMulti()
    elif EXPR == "preprocessing":
        plotPreprocessing()
    else:
        print("Unknown experiment: choose between the available options")

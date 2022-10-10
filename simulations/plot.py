#!/usr/bin/env python3
import argparse

import matplotlib.lines as mlines
import matplotlib.patches as mpatches
import matplotlib.font_manager as font_manager
import matplotlib.pyplot as plt

from utils import *

resultFolder = "final_results/"

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

def parseSpiral():
    base_name = resultFolder + 'spiral_'
    file_type = '.json'
    sizes = ["kib", "mib", "gib"]
    sizes_to_key = {"kib": 1 << 13, "mib": 1 << 23, "gib": 1 << 33}

    cpuTable = {1 << 13: 0, 1 << 23: 0, 1 << 33: 0}
    bwTable = {1 << 13: 0, 1 << 23: 0, 1 << 33: 0}
    digestTable = {1 << 13: 0, 1 << 23: 0, 1 << 33: 0}

    for s in sizes:
        f = open(base_name + s + file_type)
        out = json.load(f)
        digestTable[sizes_to_key[s]] = out["param_sz"]/1024.0 # store in KiB
        bwTable[sizes_to_key[s]] = (out["query_sz"] + out["resp_sz"])/1024.0 # store in KiB
        cpuTable[sizes_to_key[s]] = out["total_us"]/1000000 * 1000 # stored in us in the original file, store in ms

    return cpuTable, bwTable, digestTable


def plotSingleRatios():
    size_to_unit = {1<<13: "1 KiB", 1<<23: "1 MiB", 1<<33: "1 GiB"}
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
    metrics_icons = ["Off ", "On ", "Time "]
    schemes = ["simplePIR.json", "computationalDH.json", "computationalLWE128.json", "computationalLWE.json"]
    names = ['DDH', 'LWE', 'LWE$^+$', 'SimplePIR']
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

    cpuTable["spiral"], bwTable["spiral"], digestTable["spiral"] = parseSpiral()
    
    # print latex table
    metrics = (digestTable, bwTable, cpuTable)
    for dbSize in size_to_unit.keys():
        print(size_to_latex[dbSize], end = " & ")
        for i, m in enumerate(metrics):
            if i == 0:
                print(metrics_icons[i] + size_to_units_latex[dbSize][i], end = " & ")
            else:
                print(" & " + metrics_icons[i] + size_to_units_latex[dbSize][i], end = " & ")
            for scheme in schemes:
                if dbSize == 1<<33 and "DH" in scheme:
                    print("N/A", end = " & ")
                    print("N/A", end = " & ") # for ratio 
                elif "simple" in scheme:
                    print(round(m[scheme][dbSize]*size_to_multipliers[dbSize][i], 2), end = " & ")
                else:
                    ratio = round(m[scheme][dbSize]/m[schemes[0]][dbSize], 1)
                    value = round(m[scheme][dbSize]*size_to_multipliers[dbSize][i], 2)
                    if value >= 1000:
                       value = sci_notation(value, 1) 
                    if ratio >=1000:
                        ratio = sci_notation(ratio, 1)
                    print(value, end = " & ")
                    if scheme == schemes[-1]:
                        print(ratio, end = "$\\times$ ")
                    else:
                        print(ratio, end = "$\\times$ & ")
            print("\\\\")
        print("\\midrule")
        print("")

def sci_notation(number, sig_fig=2):
    ret_string = "{0:.{1:d}e}".format(number, sig_fig)
    a, b = ret_string.split("e")
    # remove leading "+" and strip leading zeros
    b = int(b)
    return "$" + a + " \\cdot 10^{" + str(b) + "}$ "
 
# def plotSingle():
#     size_to_unit = {1<<13: "1 KiB", 1<<23: "1 MiB", 1<<33: "1 GiB"}
#     base_latex = "\\multirow{3}{*}"
#     size_to_latex = {
#             1 << 13: base_latex + "{1 KiB}",
#             1 << 23: base_latex + "{1 MiB}",
#             1 << 33: base_latex + "{1 GiB}",
#         }
#     size_to_units_latex = {
#             1 << 13: ["[KiB]", "[KiB]", "[ms]"],
#             1 << 23: ["[MiB]", "[KiB]", "[ms]"],
#             1 << 33: ["[MiB]", "[KiB]", "[s]"],
#         }
#     size_to_multipliers = {
#             1 << 13: [1.0, 1.0, 1.0],
#             1 << 23: [1/1024.0, 1.0, 1.0],
#             1 << 33: [1/1024.0, 1.0, 1/1000.0],
#         }
#     metrics_icons = ["Off ", "On ", "Time "]
#     schemes = ["computationalDH.json", "computationalLWE128.json", "computationalLWE.json", "simplePIR.json"]
#     names = ['DDH', 'LWE', 'LWE$^+$', 'SimplePIR']
#     cpuTable = {}
#     bwTable = {}
#     digestTable = {}
#     for i, scheme in enumerate(schemes):
#         if scheme == "spiral":
#             continue
#         stats = allStats(resultFolder + scheme)
#         cpuTable[scheme] = {}
#         bwTable[scheme] = {}
#         digestTable[scheme] = {}
#         for j, dbSize in enumerate(sorted(stats.keys())):
#             bw = bwMean(stats, dbSize)
#             cpu = cpuMean(stats, dbSize)
#             cpuTable[scheme][dbSize] = cpu*1000*1000 # store in ms (already divided by 1000 in function)
#             bwTable[scheme][dbSize] = (bw/1024.0) # KiB, since everything is already in bytes
#             digestTable[scheme][dbSize] = stats[dbSize]['digest']/1024.0 # KiB, since everything is already in bytes
#
#     cpuTable["spiral"], bwTable["spiral"], digestTable["spiral"] = parseSpiral()
#
#     # print latex table
#     metrics = (digestTable, bwTable, cpuTable)
#     for dbSize in size_to_unit.keys():
#         print(size_to_latex[dbSize], end = " & ")
#         for i, m in enumerate(metrics):
#             if i == 0:
#                 print(metrics_icons[i] + size_to_units_latex[dbSize][i], end = " & ")
#             else:
#                 print(" & " + metrics_icons[i] + size_to_units_latex[dbSize][i], end = " & ")
#             for scheme in schemes:
#                 if dbSize == 1<<33 and "DH" in scheme:
#                     print("N/A", end = " & ")
#                 else:
#                     print(round(m[scheme][dbSize]*size_to_multipliers[dbSize][i], 2), end = " & ")
#             # print ratio
#             ratio = m[schemes[-2]][dbSize]/m[schemes[-1]][dbSize]
#             print(round(ratio, 2), end = "$\\times$ ")
#             print("\\\\")
#         print("\\midrule")
#         print("")


def plotSingle():
    font_prop = font_manager.FontProperties(family='serif', size=10)
    size_to_unit = {1<<13: "1 KiB", 1<<23: "1 MiB", 1<<33: "1 GiB"}
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
    metrics_icons = ["Off ", "On ", "Time "]
    schemes = ["computationalDH.json", "computationalLWE128.json", "computationalLWE.json", "simplePIR.json"]
    names = ['DDH', 'LWE', 'LWE$^+$', 'SimplePIR']
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

    cpuTable["spiral"], bwTable["spiral"], digestTable["spiral"] = parseSpiral()
    
    # print latex table
    metrics = (digestTable, bwTable, cpuTable)
    for dbSize in size_to_unit.keys():
        print(size_to_latex[dbSize], end = " & ")
        for i, m in enumerate(metrics):
            if i == 0:
                print(metrics_icons[i] + size_to_units_latex[dbSize][i], end = " & ")
            else:
                print(" & " + metrics_icons[i] + size_to_units_latex[dbSize][i], end = " & ")
            for scheme in schemes:
                if dbSize == 1<<33 and "DH" in scheme:
                    print("N/A", end = " & ")
                else:
                    print(round(m[scheme][dbSize]*size_to_multipliers[dbSize][i], 2), end = " & ")
            # print ratio
            ratio = m[schemes[-2]][dbSize]/m[schemes[-1]][dbSize]
            print(round(ratio, 2), end = "$\\times$ ")
            print("\\\\")
        print("\\midrule")
        print("")

    cpuData = [list(cpuTable[schemes[i]].values()) for i in range(len((schemes)))]
    cpuData[0].append(0) # bigger db for DH is NA

    offData = [list(digestTable[schemes[i]].values()) for i in range(len(schemes))]
    offData[0].append(0) # bigger db for DH is NA

    onData = [list(bwTable[schemes[i]].values()) for i in range(len(schemes))]
    onData[0].append(0) # bigger db for 

    # plot numerical result
    fig, (ax1, ax2, ax3) = plt.subplots(1, 3)
    width = 0.2
    x = np.arange(len(list(size_to_unit.values())))
     
    colors=['darkgray','gray','dimgray','black']
    for i in range(len(schemes)):
        ax1.bar(x + i*width, cpuData[i], width, color=colors[i], label=names[i]) # , color='#000080', label='Case-1', yerr=data_std[:,0])
        ax2.bar(x + i*width, offData[i], width, color=colors[i]) # , color='#000080', label='Case-1', yerr=data_std[:,0])
        ax3.bar(x + i*width, onData[i], width, color=colors[i]) # , color='#000080', label='Case-1', yerr=data_std[:,0])
    
    fig.legend(names, bbox_to_anchor=(0.14, 1, 0.8, .102), loc='lower left',
           ncol=len(names), mode="expand", borderaxespad=0., prop=font_prop, fontsize=11)

    axs = [ax1, ax2, ax3]
    for ax in axs:
        ax.set_xticks(x + width + width/2, fontproperties=font_prop)
        ax.set_xticklabels(list(size_to_unit.values()), fontproperties=font_prop)
        ax.set_xlabel('Database size', fontproperties=font_prop)
        # Axis label slightly bigger
        ax.xaxis.label.set_size(11)
        ax.set_yscale('log')
    ax1.set_ylabel('User time [ms]', fontproperties=font_prop)
    ax1.yaxis.label.set_size(11)
    ax2.set_ylabel('Offline bandwidth [KiB]', fontproperties=font_prop)
    ax2.yaxis.label.set_size(11)
    ax3.set_ylabel('Online bandwidth [KiB]', fontproperties=font_prop)
    ax3.yaxis.label.set_size(11)

    # for item in ([ax.xaxis.label, ax.yaxis.label] +
    #              ax.get_xticklabels() + ax.get_yticklabels()):
    #     item.set_fontsize(20)
    plt.tight_layout(h_pad=1.5)
    plt.savefig('figures/single_bar_multi.eps', format='eps', dpi=300, transparent=True, bbox_inches="tight")

    # plot ratio
    fig, (ax1, ax2, ax3) = plt.subplots(1, 3)
    x = np.arange(len(list(size_to_unit.values())))
     
    colors=["", 'gray', 'dimgray']
    for i in range(len(schemes)-1):
        if i != 0:
            ax1.bar(x + i*width, np.array(cpuData[i])/np.array(cpuData[-1]), width, color=colors[i], label=names[i]) # , color='#000080', label='Case-1', yerr=data_std[:,0])
            ax2.bar(x + i*width, np.array(offData[i])/np.array(offData[-1]), width, color=colors[i]) # , color='#000080', label='Case-1', yerr=data_std[:,0])
            ax3.bar(x + i*width, np.array(onData[i])/np.array(onData[-1]), width, color=colors[i]) # , color='#000080', label='Case-1', yerr=data_std[:,0])
   
    #namesThree = names.pop()
    #fig.legend(names[1:], bbox_to_anchor=(0., 1.02, 1., .102), loc='lower left',
           #ncol=len(names), mode="expand", borderaxespad=0.)

    fig.legend()

    axs = [ax1, ax2, ax3]
    for ax in axs:
        ax.set_xticks(x + width + width/2)
        ax.set_xticklabels(list(size_to_unit.values()))
        ax.set_xlabel('Database size')
        #ax.set_yscale('log')
    ax1.set_ylabel('User time ratio')
    ax2.set_ylabel('Offline bandwidth ratio')
    ax3.set_ylabel('Online bandwidth ratio')
    plt.tight_layout(h_pad=1.5)
    plt.savefig('figures/single_bar_multi_ratio.eps', format='eps', dpi=300, transparent=True, bbox_inches="tight")

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

    # cosmetics
    plt.xticks(range(len(results)), [int(x/GiB) for x in sorted(results.keys())])
    plt.ylabel('CPU time [s]')
    plt.xlabel('Database size [GiB]')
    plt.tight_layout(h_pad=1.5)
    plt.savefig('figures/preprocessing.eps', format='eps', dpi=300, transparent=True)

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
    #prepare_for_latex()
    if EXPR == "single":
        plotSingle()
        plotSingleRatios()
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

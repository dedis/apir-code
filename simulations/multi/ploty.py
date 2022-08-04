import numpy as np
import matplotlib.lines as mlines
import matplotlib.patches as mpatches
import matplotlib.pyplot as plt

from utils import *

resultFolder = "final_results/"
#resultFolder = "results/"

# styles
markers = ['.', '*', 'd', 's']
linestyles = ['-', '--', ':', '-.']
patterns = ['', '//', '.']

def statistic(a):
    return np.median(a)

def parseClientLog(file):
    tm, bw = [], []
    with open(file, "r") as f:
        for line in f:
            if "stats" in line:
                # parse log
                stats = line.replace("\n", "").partition("stats,")[2].split(",")
                bw.append(float(stats[1]))
                tm.append(float(stats[2]))

    return bw, tm

def parseServerLog(file):
    bw = []
    with open(file, "r") as f:
        for line in f:
            if "stats" in line:
                # parse log
                stats = line.replace("\n", "").partition("stats,")[2].split(",")
                bw.append(float(stats[0]))

    return bw

def plotPointMulti():
    num_servers = 5
    schemes = ["pir_classic_multi", "pir_merkle_multi"] 
    scheme_labels = ["Unauthenticated", "Authenticated"]
    client = "client_pir_classic_multi_" 

    prepare_for_latex()
    fig, axs = plt.subplots(2, sharex=True)
    
    time = []
    bandwidth = []
    for i, scheme in enumerate(schemes):
        time.append([])
        bandwidth.append([])
        for j in range(2,num_servers+1):
            # parse client logs
            client_log = "client_" + scheme + "_" + str(j) + ".log"   
            #print("parsing", client_log)
            bw, tm = parseClientLog(resultFolder + client_log)

            # parse servers log
            for k in range(0, j):
                server_log = "server_" + str(k) + "_" + scheme + "_" + str(j) + ".log"
                #print("parsing", server_log)
                bw = [a + b for a,b in zip(bw, parseServerLog(resultFolder + server_log))]
             
            time[i].append(statistic(tm))
            bandwidth[i].append(statistic(bw))

        axs[0].plot(
                [x for x in range(2,num_servers+1)],
                time[i], 
                color='black', 
                marker=markers[int(i / (len(schemes) / 2))],
                linestyle=linestyles[int(i / (len(schemes) / 2))],
                label=scheme_labels[int(i / (len(schemes) / 2))],
                linewidth=0.5,
        )
        axs[1].plot(
                [x for x in range(2,num_servers+1)],
                [x*1e-6 for x in bandwidth[i]],
                color='black', 
                marker=markers[int(i / (len(schemes) / 2))],
                linestyle=linestyles[int(i / (len(schemes) / 2))],
                label=scheme_labels[int(i / (len(schemes) / 2))],
                linewidth=0.5,
        )

    # cosmetics
    axs[0].set_ylabel('User time [s]')
    axs[0].set_xticks([x for x in range(2,num_servers+1)]), 
    axs[1].set_ylabel('Bandwidth [MiB]')
    axs[1].set_xlabel('Number of servers')
    axs[0].legend(bbox_to_anchor=(0., 1.02, 1., .102), loc='lower left',
           ncol=2, mode="expand", borderaxespad=0.)

    plt.tight_layout(h_pad=1.5)
    plt.savefig('figures/multi.eps', format='eps', dpi=300, transparent=True)

plotPointMulti()

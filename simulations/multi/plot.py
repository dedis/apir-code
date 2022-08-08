import numpy as np
import matplotlib.lines as mlines
import matplotlib.patches as mpatches
import matplotlib.pyplot as plt
import tomli

from utils import *

#resultFolder = "final_results/"
resultFolder = "results/"

# styles
markers = ['.', '*', 'd', 's']
linestyles = ['-', '--', ':', '-.']
patterns = ['', '//', '.']

# constants
GiB = 8589934592

def load_config(file):
    with open(file, 'rb') as f:
        return tomli.load(f)

def statistic(a):
    return np.mean(a)

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

def getClientStats(scheme, key):
    # parse client logs
    client_log = "client_" + scheme + "_" + str(key) + ".log"   
    return parseClientLog(resultFolder + client_log)

def getServerStats(server_id, scheme, key):
    server_log = "server_" + str(server_id) + "_" + scheme + "_" + str(key) + ".log"
    return parseServerLog(resultFolder + server_log)

    #bw = [a + b for a,b in zip(bw, parseServerLog(resultFolder + server_log))]

def plotComplex():
    schemes = ["fss_classic", "fss_auth"]
    config = load_config("fss_classic.toml")
    scheme_labels = ["Unauthenticated", "Authenticated"]

    width = 0.35  # the width of the bars
    fig, ax = plt.subplots()

    input_sizes = config['InputSizes']
    time = []
    bandwidth = []
    for i, scheme in enumerate(schemes):
        print(scheme)
        time.append([])
        bandwidth.append([])
        for input_size in input_sizes:
            bw, tm = getClientStats(scheme, input_size)

            for k in range(0, 2):
                bw = [a + b for a,b in zip(bw, getServerStats(k, scheme, input_size))]

            time[i].append(statistic(tm))
            bandwidth[i].append(statistic(bw))

    ratio_time = [time[1][i]/time[0][i] for i in range(len(time[0]))]
    ratio_bw = [bandwidth[1][i]/bandwidth[0][i] for i in range(len(time[0]))]

    x = np.arange(len(ratio_time))
    
    rects1 = ax.bar(
            x - width/2, 
            ratio_time, width, 
            label='User time', 
            color='0.3', 
            edgecolor='black', 
            )
    rects2 = ax.bar(
            x + width/2, 
            ratio_bw, width, 
            label='Bandwidth',
            color='0.7', 
            edgecolor = 'black',
            )
    ax.axhline(y = 1, color ='black', linestyle = '--')

    # cosmetics
    ax.set_ylabel('Relative overhead between \n authenticated and unauthenticated PIR')
    ax.set_xticks(x, [x for x in sorted(input_sizes)])
    ax.set_xlabel('Function-secret-sharing input size [bytes]')
    ax.set_ylim(bottom=0.9)
    ax.legend(bbox_to_anchor=(0., 1.02, 1., .102), loc='lower left',
           ncol=2, mode="expand", borderaxespad=0.)

    plt.tight_layout()
    #plt.show()
    plt.savefig('figures/complex_bars.eps', format='eps', dpi=300, transparent=True)

def plotPoint():
    schemes = ["pir_classic", "pir_merkle"]
    config = load_config("simul.toml")
    scheme_labels = ["Unauthenticated", "Authenticated"]

    fig, axs = plt.subplots(2, sharex=True)

    db_lengths = config['DBBitLengths']
    time = []
    bandwidth = []
    for i, scheme in enumerate(schemes):
        time.append([])
        bandwidth.append([])
        for dl in db_lengths:
            bw, tm = getClientStats(scheme, dl)

            # parse servers log, in this case we have only 
            # two servers
            for k in range(0, 2):
                bw = [a + b for a,b in zip(bw, getServerStats(k, scheme, dl))]
             
            time[i].append(statistic(tm))
            bandwidth[i].append(statistic(bw))

        axs[0].plot(
                [int(x/GiB) for x in db_lengths],
                time[i], 
                color='black', 
                marker=markers[int(i / (len(schemes) / 2))],
                linestyle=linestyles[int(i / (len(schemes) / 2))],
                label=scheme_labels[int(i / (len(schemes) / 2))],
                linewidth=0.5,
        )
        axs[1].plot(
                [int(x/GiB) for x in db_lengths],
                [x*1e-6 for x in bandwidth[i]],
                color='black', 
                marker=markers[int(i / (len(schemes) / 2))],
                linestyle=linestyles[int(i / (len(schemes) / 2))],
                label=scheme_labels[int(i / (len(schemes) / 2))],
                linewidth=0.5,
        )

    # cosmetics
    axs[0].set_ylabel('User time [s]')
    axs[0].set_xticks([int(x/GiB) for x in db_lengths]), 
    axs[1].set_ylabel('Bandwidth [MiB]')
    axs[1].set_xlabel('Database size [GiB]')
    axs[0].legend(bbox_to_anchor=(0., 1.02, 1., .102), loc='lower left',
           ncol=2, mode="expand", borderaxespad=0.)

    plt.tight_layout(h_pad=1.5)
    plt.savefig('figures/point.eps', format='eps', dpi=300, transparent=True)


def plotPointMulti():
    num_servers = 5
    schemes = ["pir_classic_multi", "pir_merkle_multi"] 
    scheme_labels = ["Unauthenticated", "Authenticated"]
    client = "client_pir_classic_multi_" 

    fig, axs = plt.subplots(2, sharex=True)
    
    time = []
    bandwidth = []
    for i, scheme in enumerate(schemes):
        time.append([])
        bandwidth.append([])
        for j in range(2,num_servers+1):
            bw, tm = getClientStats(scheme, j)

            # parse servers log
            for k in range(0, j):
                bw = [a + b for a,b in zip(bw, getServerStats(k, scheme, j))]
             
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

## plots
prepare_for_latex()

#plotPoint()
#plotPointMulti()
plotComplex()

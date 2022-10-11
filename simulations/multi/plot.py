import numpy as np
import matplotlib.lines as mlines
import matplotlib.patches as mpatches
import matplotlib.pyplot as plt
import tomli

from utils import *

resultFolder = "final_results/"
#resultFolder = "results/"

# styles
markers = ['.', '*', 'd', 's']
linestyles = ['-', '--', ':', '-.']
patterns = ['', '//', '.']

# constants
GiB = 1 << 33

def load_config(file):
    with open(file, 'rb') as f:
        return tomli.load(f)

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

def getClientStats(scheme, key):
    # parse client logs
    client_log = "client_" + scheme + "_" + str(key) + ".log"   
    return parseClientLog(resultFolder + client_log)

def getServerStats(server_id, scheme, key):
    server_log = "server_" + str(server_id) + "_" + scheme + "_" + str(key) + ".log"
    return parseServerLog(resultFolder + server_log)

def plotComplex():
    schemes = ["fss_classic", "fss_auth"]
    config = load_config("fss_classic.toml")
    scheme_labels = ["Unauthenticated", "Authenticated"]

    fig, axs = plt.subplots(2, sharex=True, sharey=True)

    input_sizes = config['InputSizes']
    time = []
    time_err = []
    bandwidth = []
    bandwidth_err = []
    for i, scheme in enumerate(schemes):
        time.append([])
        bandwidth.append([])
        time_err.append([])
        bandwidth_err.append([])
        for input_size in input_sizes:
            bw, tm = getClientStats(scheme, input_size)

            for k in range(0, 2):
                bw = [a + b for a,b in zip(bw, getServerStats(k, scheme, input_size))]

            time[i].append(tm)
            bandwidth[i].append(bw)

    time_np = np.array(time)
    bandwidth_np = np.array(bandwidth)

    ratio_time = np.array([statistic(time_np[1][i]/time_np[0][i]) for i in range(len(time[0]))])
    ratio_bw = np.array([statistic(bandwidth_np[1][i]/bandwidth_np[0][i]) for i in range(len(time[0]))])

    print("predicate queries max ratio user-time:", max(ratio_time))
    print("predicate queries max ratio bandwidth:", max(ratio_bw))

    err_time = np.array([np.std(time_np[1][i]/time_np[0][i]) for i in range(len(time[0]))])
    err_bw = np.array([np.std(bandwidth_np[1][i]/bandwidth_np[0][i]) for i in range(len(time[0]))])

    axs[0].plot(
            input_sizes,
            ratio_time, 
            color='black', 
            marker='x',
            markersize=2,
            linestyle=linestyles[0],
            linewidth=0.5,
    )
    axs[0].fill_between(
            input_sizes,
            np.array(ratio_time) - np.array(err_time), 
            np.array(ratio_time) + np.array(err_time), 
            color='grey',
    )
    axs[0].axhline(y = 1, color ='black', linestyle = '--', linewidth=0.5)
    axs[1].plot(
            input_sizes,
            ratio_bw,
            color='black', 
            marker='x',
            markersize=2,
            linestyle=linestyles[0],
            linewidth=0.5,
    )
    axs[1].fill_between(
            input_sizes,
            np.array(ratio_bw) - np.array(err_bw), 
            np.array(ratio_bw) + np.array(err_bw), 
            color='grey',
    )
    axs[1].axhline(y = 1, color ='black', linestyle = '--', linewidth=0.5)

    time_np = np.array(time)
    bandwidth_np = np.array(bandwidth)
    ratio_time = [statistic(time_np[1][i]/time_np[0][i]) for i in range(len(time[0]))]
    ratio_bw = [statistic(bandwidth_np[1][i]/bandwidth_np[0][i]) for i in range(len(time[0]))]

    # cosmetics
    axs[0].set_ylabel('User-time ratio')
    axs[0].set_xticks(input_sizes), 
    axs[1].set_ylabel('Bandwidth ratio')
    axs[1].set_xlabel("Length of the parameter's hidden predicate $s$ [B]")

    plt.tight_layout(h_pad=1.5)
    plt.savefig('figures/complex_lines.eps', format='eps', dpi=300, transparent=True)

    
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

    time_np = np.array(time)
    bandwidth_np = np.array(bandwidth)
    ratio_time = [statistic(time_np[1][i]/time_np[0][i]) for i in range(len(time[0]))]
    ratio_bw = [statistic(bandwidth_np[1][i]/bandwidth_np[0][i]) for i in range(len(time[0]))]

    print("point queries max ratio user-time:", max(ratio_time))
    print("point queries max ratio bandwidth:", max(ratio_bw))

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
# plotPoint()
# plotPointMulti()
plotComplex()

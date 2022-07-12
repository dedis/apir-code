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

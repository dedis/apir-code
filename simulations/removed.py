# def plotVpirBenchmarksBarBw():
#     schemes = ["vpirSingleVector.json", "vpirMultiVector.json", "vpirMultiVectorBlock.json"]
#     labels = ["Single-bit (§4.1)", "Multi-bit (§4.3)", "Multi-bit Block (§4.3)"]
#
#     Xs = np.arange(len(schemes))
#     width = 0.35
#     Ys, Yerr = [], []
#     for scheme in schemes:
#         stats = allStats(resultFolder + scheme)
#         largestDbSize = sorted(stats.keys())[-1]
#         Ys.append(stats[largestDbSize]['client']['cpu']['mean'] + stats[largestDbSize]['server']['cpu']['mean'])
#         Yerr.append(stats[largestDbSize]['client']['cpu']['std'] + stats[largestDbSize]['server']['cpu']['std'])
#
#     plt.style.use('grayscale')
#     fig, ax1 = plt.subplots()
#     color = 'black'
#     ax1.set_ylabel("CPU time [ms]", color=color)
#     ax1.tick_params(axis='y', labelcolor=color)
#     ax1.set_xticks(Xs + width / 2)
#     ax1.set_xticklabels(labels)
#     ax1.bar(Xs, Ys, width, label="CPU", color=color, yerr=Yerr)
#     plt.yscale('log')
#
#     Ys, Yerr = [], []
#     for scheme in schemes:
#         stats = allStats(resultFolder + scheme)
#         largestDbSize = sorted(stats.keys())[-1]
#         Ys.append(
#             stats[largestDbSize]['client']['bw']['mean'] / KB + stats[largestDbSize]['server']['bw']['mean'] / KB)
#         Yerr.append(
#             stats[largestDbSize]['client']['bw']['std'] / KB + stats[largestDbSize]['server']['bw']['std'] / KB)
#
#     color = 'grey'
#     ax2 = ax1.twinx()  # instantiate a second axes that shares the same x-axis
#     ax2.set_ylabel("Bandwidth [KB]")
#     ax2.bar(Xs + width, Ys, width, label="Bandwidth", color=color, yerr=Yerr)
#     ax2.legend(loc=5, fontsize=12)
#
#     # fig.tight_layout()  # otherwise the right y-label is slightly clipped
#     plt.yscale('log')
#     plt.title("Retrieval of 256B of data from 125KB DB")
#     plt.savefig('cpu_bw.eps', format='eps', dpi=300)
#     # plt.show()

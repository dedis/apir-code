#!/usr/bin/python3
import matplotlib as mpl

def prepare_for_latex():
    # parameters for Latex
    fig_width = 3.39
    golden_mean = (sqrt(5)-1.0)/2.0    # aesthetic ratio
    fig_height = fig_width*golden_mean # height in inches
    MAX_HEIGHT_INCHES = 8.0
    if fig_height > MAX_HEIGHT_INCHES:
        print("WARNING: fig_height too large:" + fig_height +
              "so will reduce to" + MAX_HEIGHT_INCHES + "inches.")
        fig_height = MAX_HEIGHT_INCHES

    params = {'backend': 'pdf', # was ps
              'text.latex.preamble': [r'\usepackage{gensymb}', r'\usepackage{sansmath}', r'\sansmath'],
              'axes.labelsize': 10, 
              'axes.titlesize': 10,
              'font.size': 10, 
              'legend.fontsize': 10, 
              'legend.loc': 'upper left',
              'lines.markersize': 10,
              'xtick.labelsize': 10,
              'ytick.labelsize': 10,
              'text.usetex': True,
              'figure.figsize': [fig_width,fig_height],
              'font.family': 'serif'
              }
    mpl.rcParams.update(params)

def allStats(file):
    client, server, total = parseResults(file)
    return stats(client), stats(server), stats(total)

def parseResults(file):
    # parse results
    results = dict()
    client = []
    server = []
    total = []
    with open(file) as f:
        data = json.load(f)
        for dbResult in data['Results']:
            if dbResult['DBLengthBits'] not in results:
                results
            client.append(0)
            server.append(0)
            for blockResult in dbResult['Results']:
                client[-1] += blockResult['Query'] + blockResult['Reconstruct']
                server[-1] += (blockResult['Answer0'] + blockResult['Answer1'])/2
            total.append(dbResult['Total'])
    return client, server, total

def stats(data):
      s = dict()
      s['mean'] = np.mean(data)
      s['std'] = np.std(data)
      return s

#!/usr/bin/python3
import numpy as np
import json
import matplotlib as mpl
import math

from collections import defaultdict


def allStats(file):
    s = dict()
    parsedResults = parseResults(file)
    for dbSize, results in parsedResults.items():
        s[dbSize] = {
            "client": stats(results["client"]),  
            "server": stats(results["server"])
        }
    return s


def parseResults(file):
    # parse results
    parsedResults = dict()
    # read json file
    with open(file) as f:
        data = json.load(f)
        # dbResults is a dict containing all the results for the
        # given size of the db, expressed in bits
        dbResults = data['Results']
        # iterate the (size, results) pairs
        for dbSize, dbResult in dbResults.items():
            client = defaultdict(list)
            server = defaultdict(list)
            # Read the results for CPU and bandwidth in each dbResult
            for measure in dbResult:
                # iterate over the repetitions of the test
                for param, repetition in measure.items():
                    # parse the results of a single block
                    for block in repetition:
                        client[param].append(block['Query'] + block['Reconstruct'])
                        server[param].append(np.mean(block['Answers']))
            parsedResults[int(dbSize)] = {"client": client, "server": server}
    return parsedResults


def stats(data):
    s = {'cpu': {}, 'bw': {}}
    s['cpu']['mean'] = np.mean(data['CPU'])
    s['cpu']['std'] = np.std(data['CPU'])
    s['bw']['mean'] = np.mean(data['Bandwidth'])
    s['bw']['std'] = np.std(data['Bandwidth'])
    return s


def prepare_for_latex():
    # parameters for Latex
    fig_width = 400
    fig_width, fig_height = set_size(fig_width)

    params = {'backend': 'ps', 
              #'text.latex.preamble': [r'\usepackage{gensymb}', r'\usepackage{sansmath}', r'\sansmath'],
              #'text.latex.preamble': [r'\usepackage{mathptmx}'],
              'axes.labelsize': 12,
              'axes.titlesize': 12,
              'font.size': 12,
              'legend.fontsize': 12,
              'lines.markersize': 8,
              'xtick.labelsize': 12,
              'ytick.labelsize': 12,
              'text.usetex': True,
              'figure.figsize': [fig_width,fig_height],
              'font.family': 'serif',
              'pgf.texsystem': 'pdflatex',
              'pgf.rcfonts': False
              }
    mpl.rcParams.update(params)


# Taken from https://jwalton.info/Embed-Publication-Matplotlib-Latex/
def set_size(width, fraction=1, subplots=(1, 1)):
    """Set figure dimensions to avoid scaling in LaTeX.

    Parameters
    ----------
    width: float or string
            Document width in points, or string of predined document type
    fraction: float, optional
            Fraction of the width which you wish the figure to occupy
    subplots: array-like, optional
            The number of rows and columns of subplots.
    Returns
    -------
    fig_dim: tuple
            Dimensions of figure in inches
    """
    if width == 'thesis':
        width_pt = 426.79135
    elif width == 'beamer':
        width_pt = 307.28987
    else:
        width_pt = width

    # Width of figure (in pts)
    fig_width_pt = width_pt * fraction
    # Convert from pt to inches
    inches_per_pt = 1 / 72.27

    # Golden ratio to set aesthetic figure height
    # https://disq.us/p/2940ij3
    golden_ratio = (5**.5 - 1) / 2

    # Figure width in inches
    fig_width_in = fig_width_pt * inches_per_pt
    # Figure height in inches
    fig_height_in = fig_width_in * golden_ratio * (subplots[0] / subplots[1])

    return fig_width_in, fig_height_in

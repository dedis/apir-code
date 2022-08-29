#!/usr/bin/python3
import numpy as np
import json
import matplotlib as mpl
import math
import re
import sys

from collections import defaultdict


def allStats(file):
    s = dict()
    parsedResults = parseResults(file)
    for dbSize, results in parsedResults.items():
        s[dbSize] = {
            "digest": results["digest"],
            "client": stats(results["client"]),  
            "server": stats(results["server"])
        }
    return s

def parseLog(file):
    stats = []
    with open(file, "r") as f:
        for line in f:
            if "stats" in line:
                # parse log
                stats.append(line.replace("\n", "").partition("stats,")[2].split(","))
    # get cores used in test
    cores = set()
    for s in stats:
        cores.add(s[0])

    # order stats by core
    statsByCores = {int(i): {} for i in cores}
    for s in stats:
        # client
        cores = int(s[0])
        if len(s) == 3:
            if 'queries' not in statsByCores[cores]:
                statsByCores[int(s[0])]['queries'] = [int(s[1])]
                statsByCores[int(s[0])]['latency'] = [float(s[2])]
            else:
                statsByCores[int(s[0])]['queries'].append(int(s[1]))
                statsByCores[int(s[0])]['latency'].append(float(s[2]))
        # server
        else:
            if 'answer' not in statsByCores[cores]:
                statsByCores[int(s[0])]['answer'] = [int(s[1])]
            else: 
                statsByCores[int(s[0])]['answer'].append(int(s[1]))
    return statsByCores

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
            digest = 0
            # Read the results for CPU and bandwidth in each dbResult
            for measure in dbResult:
                # iterate over the repetitions of the test
                for param, repetition in measure.items():
                    # parse the results of a single block
                    if param == "Digest":
                        digest = repetition
                        continue
                    for block in repetition:
                        client[param].append(block['Query'] + block['Reconstruct'])
                        # sum the values for all the servers
                        server[param].append(np.sum(block['Answers']))
            parsedResults[int(dbSize)] = {"digest": digest, "client": client, "server": server}
    return parsedResults


def stats(data):
    s = {'cpu': {}, 'bw': {}}
    s['cpu']['mean'] = np.median(data['CPU'])
    s['cpu']['std'] = np.std(data['CPU'])
    s['bw']['mean'] = np.median(data['Bandwidth'])
    s['bw']['std'] = np.std(data['Bandwidth'])
    return s

def meanFromDict(data):
    stats = dict()
    for d in data:
        stats[d]= np.median(data[d])
    return stats

def prepare_for_latex():
    # parameters for Latex
    fig_width = 241.02039
    fig_width, fig_height = set_size(fig_width)

    params = {'backend': 'ps', 
              #'text.latex.preamble': [r'\usepackage{gensymb}', r'\usepackage{sansmath}', r'\sansmath'],
              #'text.latex.preamble': [r'\usepackage{mathptmx}'],
              'axes.labelsize': 8,
              'axes.titlesize': 8,
              'font.size': 8,
              'legend.fontsize': 8,
              'lines.markersize': 5,
              'xtick.labelsize': 8,
              'ytick.labelsize': 8,
              'text.usetex': True,
              'figure.figsize': [fig_width, fig_height],
              'font.family': 'serif',
              'font.serif': 'Times',
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

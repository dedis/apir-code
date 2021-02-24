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
    fig_width = 3.39
    golden_mean = (math.sqrt(5)-1.0)/2.0    # aesthetic ratio
    fig_height = fig_width*golden_mean  # height in inches
    MAX_HEIGHT_INCHES = 8.0
    if fig_height > MAX_HEIGHT_INCHES:
        print("WARNING: fig_height too large:" + fig_height +
              "so will reduce to" + MAX_HEIGHT_INCHES + "inches.")
        fig_height = MAX_HEIGHT_INCHES

    params = {'backend': 'ps', 
              #'text.latex.preamble': [r'\usepackage{gensymb}', r'\usepackage{sansmath}', r'\sansmath'],
              #'text.latex.preamble': [r'\usepackage{mathptmx}'],
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

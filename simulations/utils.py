#!/usr/bin/python3
import numpy as np
import json

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

# Authenticated PIR
**WARNING**: This software is **not** production-ready 
and it might contain security vulnerabilities.

This code accompanies the paper "Verifiable private information retrieval"
by Simone Colombo, Kirill Nikitin, 
Bryan Ford, David J. Wu and Henry Corrigan-Gibbs, to appear at USENIX Security
2023.

This repository contains the code for multi-server and
single-server authenticated-PIR schemes and the code
for the proof of concept application Keyd, 
a privacy-preserving PGP public keys directory based on multi-server 
authenticated PIR.


# Overview

# Setup
To run the code in this repository
install [Go](https://go.dev/) (tested with Go 1.17.5)
and a C compiler (tested with GCC 12.2.0).

To reproduce the evaluation results, install 
[Python 3](https://www.python.org/downloads/), 
[Fabric](https://www.fabfile.org/),
[NumPy](https://numpy.org/) and 
[Matplotlib](https://matplotlib.org/).

We obtain our evaluation results 
on machines equipped with two
Intel Xeon E5-2680 v3 (Haswell) CPUs, each with 12 cores, 24 threads,
and operating at 2.5 GHz. Each machine has 256 GB of RAM, and
runs Ubuntu 20.04 and Go 1.17.5.
However, the code runs on any machine equipped with the 
softwares listed above.

If the machine do not support one or more of the
`-march=native`, `-msse4.1`, `-maes`, `-mavx2` or `-mavx` C compiler flags,
it is possible to remove the appropriate flags from
`lib/matrix/matrix128.go` and `lib/matrix/matrix.go`. 
Any flag modification is likely to negatively impact performance.

# Usage and experiments

## Correctness tests
To run all basic correctness tests, execute
`go test`
This command prints performance measurements to stdout.
The entire test suite takes about 6 minutes to run and it should terminate with a `PASS`,
indicating that all tests have passed.

## Multi-server point and complex queries
The code for the experiments on our multi-server authenticated-PIR schemes
is in [`simulations/multi`](simulations/multi).

To run the simulation, first modify
[`simulations/multi/config.toml`](simulations/multi/config.toml)
to indicate the IP address of the client machines and the IP addresses and
ports of the five servers machines. One can safely use the default 
port numbers that we indicate in the `simulations/multi/config.toml` file.

The [`simulations/multi/simul.toml`](simulations/multi/simul.toml) 
file contains the databases sizes, 
the number of repetitions for a single experiment and the amount of data to 
retrieve from the database. To reproduce the results of the paper, 
do not modify this file; to speed up the simulation, or to run on machines with 
insufficient RAM, one can reduce the sizes of the databases and/or the number of
repetitions.

## Single-server point queries
The code for the experiments on our single-server authenticated-PIR
resides in [`simulations`](simulations).

The experiments for single-serve schemes run on a single machine 
give the sequential nature of the protocol. 

As in the multi-server case, 
the [`simulations/multi/simul.toml`](simulations/multi/simul.toml) 
file contains the databases sizes, 
the number of repetitions for a single experiment and the amount of data to 
retrieve from the database. These can be modified to speed up the experiments
and/or use a machine with less RAM.


## Keyd: privacy-preserving key server

# Keyd
Keyd is a privacy-preserving PGP public keys directory based on multi-server
authenticated PIR.
Keyd enable a client to privately retrieve a PGP key given an email address and
to private compute several statistics over the set of PGP keys.

## WARNING
The original Go client and this [website](https://keyd.org/) are **not** production-ready software
and they are full of security vulnerabilities.
In particular, the queries issued from this website are sent to the servers in plaintext.
The queries sent from the original Go client use hard-coded secret keys
available on Github and
should **not** be considered as private.
The original Go client and this website are only a proof-of-concept for the sake
of demonstrating the performance of Keyd and they should **not** be used for
security-critical applications.

## Introduction
Keyd is a PGP public-key directory that offers
(1) classic key look-ups and
(2) computation of statistics over keys.
We implement Keyd in the two server model, where the security
properties hold as long as at least one server is honest.

Keyd servers a snapshot of SKS PGP key directory that we downloaded on 24
January 2021. We removed all the public keys larger than 8 KiB, because we
found that this was enough to include all keys that did not include large
attachments. We also removed all keys that had been revoked, keys with an
invalid format, and keys that had no email address in their metadata.
We also removed the subkeys of each public key, leaving only the primary key.
If a key included multiple emails, we indexed this key by using the primary
email. As a result, Keyd servers a total of 3,557,164 unique PGP keys.

We provide two ways to use Keyd by querying the two servers holding an exact
replica of the database.

## Go client
The code for the client is
available at [cmd/grpc/client/interactive](cmd/grpc/client/interactive) 
and can be installed from source using `go install`.

## Website frontend
This website is a frontened for the Keyd client introduced above.
The queries issued through this website are sent in cleartext to a server, which
act as a Keyd client and issue the real verifiable-PIR queries to the servers.
The answers from the servers are sent to the server simulating the client, which
executes the reconstruction procedure and forward the result to be presented on
this website.

# Citation

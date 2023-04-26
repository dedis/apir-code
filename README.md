# Authenticated PIR
**WARNING**: This software is **not** production-ready 
and it might contain security vulnerabilities.

This code accompanies the paper 
["Verifiable private information retrieval"](https://eprint.iacr.org/2023/297)
by Simone Colombo, 
Kirill Nikitin, 
Henry Corrigan-Gibbs,
David J. Wu
and Bryan Ford, to appear at USENIX Security 2023.

This repository contains the code for multi-server and
single-server authenticated-PIR schemes and the code
for the proof of concept application Keyd, 
a privacy-preserving PGP public keys directory based on multi-server 
authenticated PIR.


# Overview
The code in this repository is organizes as follows:

* [lib/client](lib/client): clients for all the authenticated and
unauthenticated PIR schemes.
* [lib/database](lib/database): databases for all the authenticated and
    unauthenticated PIR schemes, except the database for the Keyd PGP key.
* [lib/ecc](lib/ecc): error correcting code (ECC) for the
    single-server authenticated-PIR scheme based on integrity authentication;
    currently, we implement a simple repetition code.
* [lib/field](lib/field): field for the multi-server scheme for complex
    queries.
* [lib/fss](lib/fss): function-secret-sharing scheme.
* [lib/matrix](lib/matrix): matrix operations for the single-server
    authenticated-PIR scheme that relies on the LWE assumption.
* [lib/merkle](lib/merkle): Merkle tree implementation.
* [lib/monitor](lib/monitor): CPU monitoring and benchmarking tools.
* [lib/pgp](lib/pgp): utilities to create the PGP key-server database for Keyd. 
* [lib/proto](lib/proto): gRPC protocol files for deployment.
* [lib/query](lib/query): queries for the multi-server authenticated scheme for
    complex queries, i.e., available privately-computed statistics.
* [lib/server](lib/server): servers for all the authenticated and
    unauthenticated PIR schemes.
* [lib/utils](lib/utils): various utilities.
* [cmd/](cmd): clients for Keyd, both local Go clients and the web front end.
* [data/](data): data, i.e., PGP keys, for Keyd.
* [scripts/](scripts): various useful scripts.

The dump of the SKS PGP key directory can be downloaded
[here](https://drive.switch.ch/index.php/s/PoJANZvf1cOGnfS). 
The `sks*` file must be placed in the `data/sks` folder.

# Setup
To run the code in this repository
install [Go](https://go.dev/) (tested with Go 1.17.5 and 1.19.5)
and a C compiler (tested with GCC 9.4.0).

To reproduce the evaluation results, install 
[GNU Make](https://www.gnu.org/software/make/),
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

<!--# Usage and experiments-->

## Correctness tests
To run all basic correctness tests, execute
`go test`
This command prints performance measurements to stdout.
The entire test suite takes about 6 minutes to run and it should terminate with a `PASS`,
indicating that all tests have passed.

<!--## Multi-server point and complex queries-->
<!--The code for the experiments on our multi-server authenticated-PIR schemes-->
<!--is in [`simulations/multi`](simulations/multi).-->

<!--To run the simulation, first modify-->
<!--[`simulations/multi/config.toml`](simulations/multi/config.toml)-->
<!--to indicate the IP address of the client machines and the IP addresses and-->
<!--ports of the five servers machines. One can safely use the default -->
<!--port numbers that we indicate in the `simulations/multi/config.toml` file.-->

<!--The [`simulations/multi/simul.toml`](simulations/multi/simul.toml) -->
<!--file contains the databases sizes, -->
<!--the number of repetitions for a single experiment and the amount of data to -->
<!--retrieve from the database. To reproduce the results of the paper, -->
<!--do not modify this file; to speed up the simulation, or to run on machines with -->
<!--insufficient RAM, one can reduce the sizes of the databases and/or the number of-->
<!--repetitions.-->

<!--TODO HERE FINISH-->

<!--The multi-server authenticated-PIR scheme -->
<!--for point queries needs database preprocessing:-->
<!--the servers compute a Merkle-->
<!--tree over the database entries along-->
<!--with their indexes.-->
<!--Then for each entry, each server constructs a Merkle proof-->
<!--of inclusion in the rooted Merkle tree and attaches this proof-->
<!--to each database record.-->
<!--We measure the CPU time that a single server takes to process the database -->
<!--with an experiment that can be executed as follows. From the root -->
<!--of the repository, run the following commands:-->
<!--```-->
<!--cd simulations-->
<!--make preprocessing-->
<!--```-->

<!--To reproduce the plot run the following command in the same directory:-->
<!--```-->
<!--python plot.py -e preprocessing-->
<!--```-->
<!--The resulting plot is saved in `figures/preprocessing.eps`.-->

<!--## Single-server point queries-->
<!--The code for the experiments on our single-server authenticated-PIR-->
<!--resides in [`simulations`](simulations).-->

<!--The experiments for single-serve schemes run on a single machine -->
<!--give the sequential nature of the protocol. -->

<!--As in the multi-server case, -->
<!--the [`simulations/multi/simul.toml`](simulations/multi/simul.toml) -->
<!--file contains the databases sizes, -->
<!--the number of repetitions for a single experiment and the amount of data to -->
<!--retrieve from the database. These can be modified to speed up the experiments-->
<!--and/or use a machine with less RAM.-->

<!--To run the single-server experiments, first clone this repository on the server. -->
<!--Form the root of repository, run the command-->
<!--```-->
<!--cd simulations-->
<!--make single-->
<!--```-->

<!--To reproduce the plots run the following commands in the same directory:-->
<!--```-->
<!--python plot.py -e single-->
<!--```-->
<!--This command saves the plot in `figures/single_bar_multi.eps` and prints a LaTeX-->
<!--table in the terminal; the table is not used in the paper but it is useful to-->
<!--extrapolate the overheads among schemes.-->

<!--## Keyd: privacy-preserving key server-->

The branch [sid](https://github.com/dedis/apir-code/tree/sid) enables to run the
tests using less physical machines than the servers used by the different
experiments. We decided not to merge this branch into the main branch because
multi-server (authenticated) PIR schemes need non-colluding, i.e., different,
servers for security.


# Citation

```
@inproceedings{colombo23authenticated,
  author    = {Simone Colombo and Kirill Nikitin and Henry Corrigan-Gibbs and David J. Wu and Bryan Ford},
  title     = {Authenticated private information retrieval},
  booktitle = {USENIX Security},
  year      = {2023}
}
```

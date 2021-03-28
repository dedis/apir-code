# Verifiable PIR
This repository contains the code for the verifiable PIR schemes and the code
for the proof of concept application, Keyd. 
Keyd is a privacy-preserving PGP public keys directory based on verifiable PIR.

### TODO
The current priorities for the implementation are:
 * Reduce memory consumption of Keyd and (if possible) improve its performance.
 * Fix the bugs in the gRPC logic to perform last part of the empirical
     evaluation for the paper.
 * If the memory consumption can be reduced to an acceptable level, setup two
     AWS instances (Lambda is not the appropriate solution) and design a client
     for Keyd.

### Verifiable PIR taxonomy

### Keyd workflow

### Code organization
The [lib](lib) folder is organized as follows.
  * [lib/client](lib/client): contains all the logic for the client operations
    in the different (verifiable) PIR schemes. The file
    [lib/client/client.go](lib/client/client.go) contains the definition of the
    Client interface and general functions used to implement the client for
    different schemes. The other files contains the types and functions specific
    to each scheme.
  * [lib/constants](lib/constants): defines the constant used by the schemes. 
  * [lib/database](lib/database): contains the database's logic. The database
      is created once by the servers  and this should not have a negative impact
      on Keyd, which uses the `DB` type defined in
      [lib/database/db.go](lib/database/db.go).
  * [lib/dpf](lib/dpf): contains the implementation of the distributed point
      function (DPF) used in combination with verifiable PIR. 
  * [lib/field](lib/field): contains the logic of the field used by the
      information-theoretic VPIR protocols.
  * [lib/merkle](lib/merkle): is the library for the Merkle tree. This VPIR
      approach is not used by Keyd and therefore this folder can be safely
      ignored.
  * [lib/monitor](lib/monitor): implements a CPU time timer.
  * [lib/pgp](lib/pgp): contains the logic to parse the SKS dump (PGP keys
      dump). 
  * [lib/proto](lib/proto): contains the gRPC logic.
  * [lib/server](lib/server): same structure as [lib/client](lib/client).
  * [lib/utils](lib/utils): contains various utils.

The [cmd](cmd) folder is organized as follows.

The file [config.toml](config.toml) contains the configurations for the
client-server logic. Right now, it contains only IP addresses and ports of the
servers.

The [data](data) folder contains the data used by Keyd. In particular, it
contains the files of the SKS dump, i.e. the set of PGP keys. These files are
available on the shared Google Drive.

The test of the (V)PIR schemes are specified



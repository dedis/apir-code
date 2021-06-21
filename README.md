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
The repository contains different verifiable PIR schemes. Keyd will be based
either on the multi-bit DPF-based scheme or on the multi-bit matrix-based
scheme. These are called `dpf` and `it` schemes in the code, respectively. In
particular, both [lib/client](lib/client) and [lib/server](lib/server) contain
files `dpf.go` and `it.go` which contains the logic for the verifiable PIR
schemes on which we build Keyd.

### Keyd workflow
Keyd works with a client and two servers. The input of the system is an email
address and the output is the public PGP key of the given email address.
The client client gets the email address and hashes is to identify the index of
the entry to be retrieved using the verifiable PIR scheme. Several email
addresses map to the same database entry (which is in fact a column of field
elements encoding PGP public keys). The client computes the verifiable PIR
queries using the index and sends the queries to the servers. For this, the
method `QueryBytes` (implemented by both DPF-based and matrix-based PIR) is
used. Upon receiving a query, the server compute the appropriate answer using
the `AnswerBytes` method and send the encoded answer to the client. 
Upon receiving the two answers the client uses the `ReconstructBytes` function
to check the correctness of the retrieved answers and reconstruct the database
entry. Finally, the client scans all the keys in the reconstructed entry and
returns the correct key with respect to the initial email address.

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
  * [lib/monitor](lib/monitor): implements a CPU time timer, used only for tests
      and evaluation.
  * [lib/pgp](lib/pgp): contains the logic to parse the SKS dump (PGP keys
      dump). 
  * [lib/proto](lib/proto): contains the gRPC logic.
  * [lib/server](lib/server): same structure as [lib/client](lib/client).
  * [lib/utils](lib/utils): contains various utils.

The [cmd](cmd) folder is organized in two distinct folders, [cmd/grpc](cmd/grpc)
and [cmd/aws](cmd/aws). The former implements the gRPC logic, the latter is to
be considered useless now, since AWS Lambda is not a viable solution to deploy
Keyd. 

The file [config.toml](config.toml) contains the configurations for the
client-server logic. Right now, it contains only IP addresses and ports of the
servers.

The [data](data) folder contains the data used by Keyd. In particular, it
contains the files of the SKS dump, i.e. the set of PGP keys. These files are
available on the shared Google Drive.

The test of the (V)PIR schemes are specified in the `*_test.go` files. 
The test relative to Keyd and its underlying cryptographic primitives are in
[key_test.go](key_test.go) and [vpir_test.go](vpir_test.go).

[scripts](scripts) and [simulations](simulations) contain useful scripts and
code for the evaluation of the cryptographic schemes. 
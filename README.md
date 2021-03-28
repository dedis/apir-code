# Verifiable PIR

### TODO

### Verifiable PIR taxonomy

### Keyd workflow

### Code organization
The 
The [lib](lib) folder is organized as follows.
  * [lib/client](lib/client): contains all the logic for the client operations
    in the different (verifiable) PIR schemes. The file
    [lib/client/client.go](lib/client/client.go) contains the definition of the
    Client interface and general functions used to implement the client for
    different schemes. The other files contains the types and functions specific
    to each scheme.
  * [lib/constants](lib/constants):
  * [lib/database](lib/database):
  * [lib/dpf](lib/dpf):
  * [lib/field](lib/field):
  * [lib/merkle](lib/merkle):
  * [lib/monitor](lib/monitor):
  * [lib/pgp](lib/pgp):
  * [lib/proto](lib/proto):
  * [lib/server](lib/server): same structure as [lib/client](lib/client).
  * [lib/utils](lib/utils):

The [cmd](cmd) folder is organized as follows.

The file [config.toml](config.toml) contains the configurations for the
client-server logic. Right now, it contains only IP addresses and ports of the
servers.

The [data](data) folder contains the data used by Keyd. In particular, it
contains the files of the SKS dump, i.e. the set of PGP keys. These files are
available on the shared Google Drive.

The test of the (V)PIR schemes are specified



# Authenticated PIR
**WARNING**: This software is **not** production-ready and is full of security
vulnerabilities.

This repository contains the code for multi-server and
single-server authenticated-PIR schemes and the code
for the proof of concept application, Keyd, 
a privacy-preserving PGP public keys directory based on multi-server 
authenticated PIR.

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


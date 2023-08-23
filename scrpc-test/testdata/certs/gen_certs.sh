#!/usr/bin/env bash

set -ex

cfssl gencert -initca server-ca-csr.json | cfssljson -bare server-ca && \
cfssl gencert -initca client1-ca-csr.json | cfssljson -bare client1-ca && \
cfssl gencert -initca client2-ca-csr.json | cfssljson -bare client2-ca

cfssl gencert \
-ca=server-ca.pem \
-ca-key=server-ca-key.pem \
-config=ca-config.json \
-profile=server \
server-csr.json | cfssljson -bare server

cfssl gencert \
-ca=client1-ca.pem \
-ca-key=client1-ca-key.pem  \
-config=ca-config.json \
-profile=client \
client1-csr.json | cfssljson -bare client1

cfssl gencert \
-ca=client2-ca.pem \
-ca-key=client2-ca-key.pem  \
-config=ca-config.json \
-profile=client \
client2-csr.json | cfssljson -bare client2

rm -f *.csr *ca-key.pem

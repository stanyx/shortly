[![Build Status](https://travis-ci.com/stanyx/shortly.svg?branch=master)](https://travis-ci.com/stanyx/shortly)

## SHORTLY

Urlshortener application, inspired by Bit.ly, TinyURL, Google

### Project goals

 - Simplicity
 - Easy deployment
 - Cloud ready

### Features included

 - Public REST API for creating, deleting links
 - Private REST API with payed plans
 - Redirect endpoint for url transformation
 - Support for 3 cache engines: In-memory, Memcached, BoltDB
 - Users and roles management (RBAC)
 - Billing functionality
 - Statisticts aggregation (not ready)
 - Prometheus monitoring integration (in progress)
 - Kubernetes ready (in progress)

### Prerequisites

Using TLS with server

    openssl genrsa -out server.key 2048
    openssl ecparam -genkey -name secp384r1 -out server.key
    openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650

### Installing from Docker

    docker pull stanyx/shortly
    docker run --expose=[port] -p [port]:[port] shortly
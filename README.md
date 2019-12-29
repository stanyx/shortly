[![Build Status](https://travis-ci.com/stanyx/shortly.svg?branch=master)](https://travis-ci.com/stanyx/shortly)

## SHORTLY

Simple project implementing urlshortener functionality, like Bit.ly, TinyURL, Google urlshorteners

### Project goals

 - Simplicity
 - Easy deployment
 - Cloud ready

### Features included

 - Best practices
 - Fully tested
 - Benchmarks
 - Public REST API for creating, deleting urls
 - Private REST API with payed plans
 - Redirect endpoint for url transformation
 - Support for 3 cache engines: Memcached, Redis, BoltDB (not ready)
 - Users and roles management (RBAC) (not ready)
 - Billing functionality (not ready)
 - Statisticts aggregation (not ready)
 - Prometheus integration (not ready)
 - Kubernetes ready (not ready)

### Prerequisites

Using TLS with server

    openssl genrsa -out server.key 2048
    openssl ecparam -genkey -name secp384r1 -out server.key
    openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650

### Installing from source

TODO

### Installing from Docker

docker pull stanyx/shortly
docker run --expose=[port] -p [port]:[port] shortly

### CONTRIBUTING

TODO
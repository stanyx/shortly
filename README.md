[![Go Report Card](https://goreportcard.com/badge/stanyx/shortly)](https://goreportcard.com/report/stanyx/shortly)
[![Build Status](https://travis-ci.com/stanyx/shortly.svg?branch=master)](https://travis-ci.com/stanyx/shortly)
[![codebeat badge](https://codebeat.co/badges/311fecd5-7eab-4c56-8edd-780e3aecb7ba)](https://codebeat.co/projects/github-com-stanyx-shortly-master)

# SHORTLY

Urlshortener application, inspired by Bit.ly, TinyURL, Google

## Project goals

 - Simplicity
 - Easy deployment
 - Cloud ready

## Features included

 - Public REST API for creating, deleting links
 - Private REST API with payed plans
 - Redirect endpoint for url transformation
 - Support for 3 cache engines: In-memory, Memcached, BoltDB
 - Users and roles management (RBAC)
 - Billing functionality with Stripe
 - Statisticts aggregation (in progress)
 - Prometheus monitoring integration (in progress)
 - Kubernetes ready (in progress)

## Prerequisites

Using TLS with server

```bash
    openssl genrsa -out server.key 2048
    openssl ecparam -genkey -name secp384r1 -out server.key
    openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
```

## Installing from Docker

```bash
    docker pull stanyx/shortly
    docker run --expose=[port] -p [port]:[port] shortly
```
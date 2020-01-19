[![Go Report Card](https://goreportcard.com/badge/stanyx/shortly)](https://goreportcard.com/report/stanyx/shortly)
[![Build Status](https://travis-ci.com/stanyx/shortly.svg?branch=master)](https://travis-ci.com/stanyx/shortly)
[![codebeat badge](https://codebeat.co/badges/311fecd5-7eab-4c56-8edd-780e3aecb7ba)](https://codebeat.co/projects/github-com-stanyx-shortly-master)

# SHORTLY

Urlshortener written in [Golang](https://golang.org) and [ReactJS](https://reactjs.org). Inspired by Bit.ly, TinyURL.

## Project category

- Pet/homebrew
- Expiremental
- Education
- Research
- Hacking

 Semi-production ready :)

## Project status

WORK IN PROGRESS

Project is at very early stage of development. That means that breaking changes may arise at any time.

### ROADMAP (2020)

 - semantic versioning
 - api stabilization
 - tests for api
 - frontend part
 - migration to microservice stack

## Why create another urlshortener application?

 Mainly for educational purposes. It's a pet/homebrew project, which you can try to create with yourself. 
 
 This project presents one of many possible approach for creating production grade scalable web application with 3 Tier paradigm.

 Junior/middle software developers may use it for self-education and skill gaining.

## How to use it
 
 Best way is forking. You may research code with your favorite code editor. Try to run and experiment with it.
 If you want to help with a tip, create an issue. Pull requests are welcome too.

## Project goals

 - Simplicity
 - Easy deployment
 - Scalable
 - Cloud ready

## Tech stack

 - Web:                           Golang/Chi router
 - Caching:                       Memcached
 - Persistent storage (Database): PostgreSQL
 - Frontend:                      React

## Features included

 - Public endpoint for shorterning links
 - Single redirect endpoint
 - Signin, signup functionality
 - HTTPS support
 - Deployable with Heroku or Docker
 - JWT (JSON Web Tokens) authentication
 - Private REST API for managing links
 - Support for 3 cache engines: In-memory, Memcached, BoltDB
 - Multitenancy, users and roles management (RBAC)
 - Billing functionality with Stripe
 - Prometheus monitoring integration
 - Statisticts aggregation (coming soon)
 - Kubernetes ready (coming soon)

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
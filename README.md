[![Go Report Card](https://goreportcard.com/badge/stanyx/shortly)](https://goreportcard.com/report/stanyx/shortly)
[![Build Status](https://travis-ci.com/stanyx/shortly.svg?branch=master)](https://travis-ci.com/stanyx/shortly)
[![codebeat badge](https://codebeat.co/badges/311fecd5-7eab-4c56-8edd-780e3aecb7ba)](https://codebeat.co/projects/github-com-stanyx-shortly-master)

# SHORTLY

URL shortener written in [Golang](https://golang.org) and [ReactJS](https://reactjs.org). Inspired by Bit.ly, TinyURL.

## Project category

- Pet/homebrew
- Expiremental
- Education
- Research
- Hacking

 Semi-production ready :)

## Project status

WORK IN PROGRESS

The project is at an very early stage of development. Breaking changes may occur at any time.

### ROADMAP (2020)

 - semantic versioning
 - api stabilization
 - tests for api
 - frontend
 - migration to the microservice stack

## Why create another URL shortener application?

 Mainly for educational purposes. It's a pet/homebrew project, like the ones you can try to create on your own. 
 
 This project represents one of the many possible approaches to creating production-grade scalable 3-tier architecture web applications.

 Junior/middle software developers may use it for self-education and gaining new skills.

## How to use it
 
 The best way is to fork it. You can take a closer look at the code in your favorite editor. Try to run it and play around with it.
 If you have a suggestion how to make the project better, create an issue. Pull requests are welcome too.

## Project goals

 - Simplicity
 - Easy deployment
 - Scalable
 - Cloud-ready

## Tech stack

 - Web:                           Golang/Chi router
 - Caching:                       Memcached
 - Persistent storage (Database): PostgreSQL
 - Frontend:                      React

## Features included

 - Public endpoint for shortening links
 - Single redirect endpoint
 - Sign-in, sign-up functionality
 - HTTPS support
 - Deployable with Heroku or Docker
 - JWT (JSON Web Tokens) authentication
 - Private REST API for managing links
 - Support for cache engines: in-memory, Memcached, BoltDB
 - Multitenancy, users and roles management (RBAC)
 - Billing functionality with Stripe
 - Prometheus monitoring integration
 - Statistics aggregation (coming soon)
 - Kubernetes-ready (coming soon)

## Prerequisites

Using TLS with server

```bash
openssl genrsa -out server.key 2048
openssl ecparam -genkey -name secp384r1 -out server.key
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
```

## Installing with Docker

```bash
docker pull stanyx/shortly
docker run --expose=[port] -p [port]:[port] shortly
```

# Service Registry [![CircleCI](https://circleci.com/gh/lodastack/registry.svg?style=svg&circle-token=d86b3917d23b1883d3b0682c877a3114eff49fc0)](https://circleci.com/gh/lodastack/registry) [![Go Report Card](https://goreportcard.com/badge/github.com/lodastack/registry)](https://goreportcard.com/report/github.com/lodastack/registry)

## Build

use `make build` build this project.

## Quick Start


### single node 

```bash

registry -config /etc/node0.conf

```

### Start a node and join into a cluster by raft address.

```bash

registry -config /tmp/node1.conf -join 127.0.0.1:9000

```
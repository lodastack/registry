# Service Registry [![CircleCI](https://circleci.com/gh/lodastack/registry.svg?style=svg&circle-token=d86b3917d23b1883d3b0682c877a3114eff49fc0)](https://circleci.com/gh/lodastack/registry)

## Build

use `make build` build this project.

## Quick Start


### single node 

```bash

registry -config /etc/node0.conf

```

### add a node to cluster

```bash

registry -join 127.0.0.1:8000 -config /tmp/node1.conf

```
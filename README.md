# service registry

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
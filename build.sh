#!/bin/bash

set -e

# build bin file
make

# make my dir
mkdir -p $BUILD_ROOT/conf $BUILD_ROOT/bin
mv cmd/registry/registry $BUILD_ROOT/bin/.
cp etc/registry.sample.conf $BUILD_ROOT/conf/.

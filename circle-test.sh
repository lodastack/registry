#!/bin/bash

make dep

go get github.com/lodastack/registry

go test -timeout 60s -v ./...
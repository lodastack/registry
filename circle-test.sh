#!/bin/bash

go version

make dep

go test -timeout 60s -v ./...
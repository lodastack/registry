#!/bin/bash

make dep

go test -timeout 60s -v ./...
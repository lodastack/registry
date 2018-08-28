#!/bin/bash

go version
export GO111MODULE=on
go mod download
go mod verify
go test -timeout 60s -v ./...
cd ./cmd/registry && go build -v -mod readonly

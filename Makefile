all: build

fmt:
	gofmt -l -w -s */

dep:fmt
	go get github.com/boltdb/bolt
	go get github.com/lodastack/log
	go get github.com/hashicorp/raft
	go get github.com/BurntSushi/toml
	go get github.com/hashicorp/raft-boltdb

build: fmt 
	cd cmd/registry && go build -v

install: fmt
	cd cmd/registry && go install

clean:
	cd cmd/registry && go clean

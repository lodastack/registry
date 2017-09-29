all: build

fmt:
	gofmt -l -w -s */

dep:fmt
	gdm restore

build: dep 
	cd cmd/registry && go build -v

install: fmt
	cd cmd/registry && go install

clean:
	cd cmd/registry && go clean

test:
	go test ./...

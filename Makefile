all: build

fmt: dep
	gofmt -l -w -s */

build: fmt
	cd cmd/registry && go build -v

install: fmt
	cd cmd/registry && go install

clean:
	cd cmd/registry && go clean

dep:
	go mod download

tidy:
	go mod tidy

test: tidy
	go test ./...

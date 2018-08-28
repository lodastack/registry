all: build

fmt:
	gofmt -l -w -s */

build: fmt
	cd cmd/registry && go build -v

install: fmt
	cd cmd/registry && go install

clean:
	cd cmd/registry && go clean

tidy:
	go mod tidy

test: tidy
	go test ./...

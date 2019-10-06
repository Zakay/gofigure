build:
	docker run --rm -v $$PWD:/app -w /app -e GOPATH=/app/gopath golang:1.10 /bin/sh -c "go get -d ./... && go build -o bin/gofigure"

build-local:
	go get -d ./...  
	go build -o bin/gofigure

run:
	docker run --rm -v $$PWD:/app -w /app -e GOPATH=/app/gopath golang:1.10 ./bin/gofigure -i files/config.define_roots.fig

run-local:
	./bin/gofigure -i files/config.fig

.PHONY: run run-local build build-local

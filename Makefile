hook:
	git config core.hooksPath .githooks

init:
	go run ./... -init

run:
	go run ./...

build:
	go build ./...

test:
	go test -v ./...

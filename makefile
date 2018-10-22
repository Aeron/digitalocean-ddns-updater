.PHONY: deps server

deps:
	go mod tidy
	go mod verify

server: deps
	go run *.go

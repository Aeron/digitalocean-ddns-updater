tidy:
    go mod tidy
    go mod verify

run: tidy
    go run ./...

test *FLAGS:
    go test ./... {{FLAGS}}

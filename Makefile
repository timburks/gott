all:
	go install ./...

lint:
	golangci-lint run

test:
	go test ./...

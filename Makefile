.PHONY: example
example:
	go run examples/basic/main.go

.PHONY: lint
lint:
	go fmt .
	go vet .
	golint .
	golangci-lint run .
	go test .

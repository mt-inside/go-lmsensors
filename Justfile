generate:
	go generate ./...

lint: generate
	gofumpt -l -w .
	goimports -local github.com/mt-inside/go-lmsensors -w .
	go vet ./...
	staticcheck ./...
	golangci-lint run ./... # Will never be able to --enable-all on code using CGO

example-basic:
	go run examples/basic/main.go

example-all-fans:
	go run examples/all-fans/main.go

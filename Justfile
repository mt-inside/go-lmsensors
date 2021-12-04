example-basic:
	go run examples/basic/main.go

example-all-fans:
	go run examples/all-fans/main.go

check:
	build/check-go

gen:
	go generate ./...

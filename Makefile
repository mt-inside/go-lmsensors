.PHONY: example-basic
example-basic:
	go run examples/basic/main.go

.PHONY: check
check:
	build/check-go

.PHONY: gen
gen:
	go generate ./...

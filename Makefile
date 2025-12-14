example:
	go run examples/example.go
test:
	go clean -testcache && go test ./... $(ARGS)


.PHONY: example test

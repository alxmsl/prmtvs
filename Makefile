.PHONY: fmt
fmt:
	gofmt -l -w `find . -type f -name '*.go' -not -path "./vendor/*"`
	goimports -l -w `find . -type f -name '*.go' -not -path "./vendor/*"`

.PHONY: test
test:
	go test ./... -v -race

%:
	@:

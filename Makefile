unit-test:
	go test -gcflags=all=-l -coverprofile=coverage.txt $(shell go list ./pkg/... ./cmd/...)

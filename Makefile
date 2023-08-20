include makefiles/dependency.mk

lint: golangci
	$(GOLANGCILINT) run ./...

staticcheck: staticchecktool
	$(STATICCHECK) ./...

fmt: goimports
	$(GOIMPORTS) -local github.com/kubevela-contrib/velaux-database-migrator -w $$(go list -f {{.Dir}} ./...)

go-check:
	go fmt ./...
	go vet ./...
mod:
	go mod tidy

reviewable: mod lint staticcheck fmt go-check

check-diff: reviewable
	git --no-pager diff
	git diff --quiet || (echo please run 'make reviewable' to include all changes && false)
	echo branch is clean
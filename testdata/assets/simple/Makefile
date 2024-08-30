u := $(if $(update),-u)

.PHONY: deps
deps:
	go get ${u}
	go mod tidy

.PHONY: devel-deps
devel-deps:
	go install github.com/Songmu/godzil/cmd/godzil@latest

.PHONY: test
test:
	go test

.PHONY: release
release: devel-deps
	godzil release

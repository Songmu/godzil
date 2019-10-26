u := $(if $(update),-u)

export GO111MODULE=on

.PHONY: deps
deps:
	go get ${u} -d
	go mod tidy

.PHONY: devel-deps
devel-deps:
	GO111MODULE=off go get ${u} \
	  golang.org/x/lint/golint            \
	  github.com/Songmu/godzil/cmd/godzil

.PHONY: test
test:
	go test

.PHONY: lint
lint: devel-deps
	golint -set_exit_status

.PHONY: release
release: devel-deps
	godzil release

VERSION = $(shell godzil show-version)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X github.com/Songmu/godzil.revision=$(CURRENT_REVISION)"
u := $(if $(update),-u)

export GO111MODULE=on

.PHONY: deps
deps:
	go get ${u} -d
	go mod tidy

.PHONY: devel-deps
devel-deps:
	GO111MODULE=off go get ${u} \
	  golang.org/x/lint/golint                  \
	  github.com/Songmu/godzil/cmd/godzil       \
	  github.com/Songmu/gocredits/cmd/gocredits \
	  github.com/tcnksm/ghr                     \
	  github.com/Songmu/statikp/cmd/statikp

.PHONY: assets
assets:
	statikp -m -src testdata/assets -dotfiles

.PHONY: test
test:
	go test

.PHONY: lint
lint: devel-deps
	golint -set_exit_status

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) ./cmd/godzil

.PHONY: install
install:
	go install -ldflags=$(BUILD_LDFLAGS) ./cmd/godzil

.PHONY: release
release: devel-deps
	godzil release

CREDITS: deps devel-deps go.sum
	gocredits -w

.PHONY: crossbuild
crossbuild: CREDITS
	godzil crossbuild -pv=v$(VERSION) -build-ldflags=$(BUILD_LDFLAGS) \
      -os=linux,darwin,windows -d=./dist/v$(VERSION) ./cmd/*

.PHONY: upload
upload:
	ghr -body="$$(godzil changelog --latest -F markdown)" v$(VERSION) dist/v$(VERSION)

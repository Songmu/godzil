CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X github.com/Songmu/godzil.revision=$(CURRENT_REVISION)"
u := $(if $(update),-u)

.PHONY: deps
deps:
	go get ${u}
	go mod tidy

.PHONY: devel-deps
devel-deps:
	go install github.com/Songmu/gocredits/cmd/gocredits@v0.5.0

.PHONY: test
test:
	go test
	make assets-test

.PHONY: assets-test
assets-test:
	make assets
	@git diff --exit-code --quiet testdata/assets || \
      (echo '💢 Inconsistency in testdata/assets' && false)

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) ./cmd/godzil

.PHONY: assets
assets:
	for profile in simple basic web; do \
      cp -r testdata/assets/_common/ testdata/assets/$$profile; \
    done

.PHONY: install
install:
	go install -ldflags=$(BUILD_LDFLAGS) ./cmd/godzil

.PHONY: prepare-release
prepare-release: devel-deps
	go mod tidy
	gocredits -w
	git update-index --add --remove -- go.mod go.sum CREDITS

CREDITS: deps devel-deps go.sum
	gocredits -w

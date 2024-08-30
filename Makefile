VERSION = $(shell ./godzil show-version)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-s -w -X github.com/Songmu/godzil.revision=$(CURRENT_REVISION)"
u := $(if $(update),-u)

.PHONY: deps
deps:
	go get ${u}
	go mod tidy

.PHONY: devel-deps
devel-deps: build
	go install github.com/tcnksm/ghr@latest

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

.PHONY: release
release: devel-deps
	./godzil release

CREDITS: deps devel-deps go.sum
	./godzil credits -w

.PHONY: crossbuild
crossbuild: CREDITS
	./godzil crossbuild -pv=v$(VERSION) -build-ldflags=$(BUILD_LDFLAGS) \
      -os=linux,darwin,windows -d=./dist/v$(VERSION) ./cmd/*

.PHONY: upload
upload:
	ghr -body="$$(./godzil changelog --latest -F markdown)" v$(VERSION) dist/v$(VERSION)

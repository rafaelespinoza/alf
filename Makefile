GO ?= go
GOSEC ?= gosec

PKG_IMPORT_PATH=github.com/rafaelespinoza/alf

build:
	$(GO) build $(PKG_IMPORT_PATH) $(ARGS)

build-examples:
	mkdir -pv bin && $(GO) build -o ./bin/full_example ./examples/full

mod:
	$(GO) mod tidy

test:
	$(GO) test $(PKG_IMPORT_PATH)/... $(ARGS)

vet:
	$(GO) vet $(PKG_IMPORT_PATH)/... $(ARGS)

# Run a security scanner over the source code. This Makefile won't install the
# scanner binary for you, so check out the gosec README for instructions:
# https://github.com/securego/gosec
#
# If necessary, specify the path to the built binary with the GOSEC env var.
#
# Also note, the package paths (last positional input to gosec command) should
# be a "relative" package path. That is, starting with a dot.
gosec:
	$(GOSEC) $(ARGS) ./...

STANDARDS_DIR ?= $(wildcard $(CURDIR)/../go-standards)
include standard.mk

.PHONY: fuzz

# The independent point model cross-checks every range algorithm.
fuzz:
	$(GO) test -run '^$$' -fuzz '^FuzzSet$$' -fuzztime=30s .

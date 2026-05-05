BUILD_DIR := build

# FetchContent places the ghostty source here.
GHOSTTY_ZIG_OUT := $(CURDIR)/$(BUILD_DIR)/_deps/ghostty-src/zig-out
PKG_CONFIG_PATH := $(GHOSTTY_ZIG_OUT)/share/pkgconfig
DYLD_LIBRARY_PATH := $(GHOSTTY_ZIG_OUT)/lib
LD_LIBRARY_PATH := $(GHOSTTY_ZIG_OUT)/lib

# Stamp file for the native libghostty-vt build.
STAMP := $(BUILD_DIR)/.ghostty-built

# Cross-compilation target definitions.
# Each entry maps a make target suffix to GOOS, GOARCH, and the zig
# target triple used both for `zig cc` (Go cgo) and for libghostty-vt
# itself via the GHOSTTY_VT_CROSS_TARGETS CMake variable.
CROSS_TARGETS := linux-amd64 linux-arm64 macos-amd64 macos-arm64 windows-amd64 windows-arm64

linux-amd64_GOOS   := linux
linux-amd64_GOARCH := amd64
linux-amd64_ZIG    := x86_64-linux-gnu

linux-arm64_GOOS   := linux
linux-arm64_GOARCH := arm64
linux-arm64_ZIG    := aarch64-linux-gnu

macos-amd64_GOOS   := darwin
macos-amd64_GOARCH := amd64
macos-amd64_ZIG    := x86_64-macos

macos-arm64_GOOS   := darwin
macos-arm64_GOARCH := arm64
macos-arm64_ZIG    := aarch64-macos

windows-amd64_GOOS   := windows
windows-amd64_GOARCH := amd64
windows-amd64_ZIG    := x86_64-windows-gnu

windows-arm64_GOOS   := windows
windows-arm64_GOARCH := arm64
windows-arm64_ZIG    := aarch64-windows-gnu

.PHONY: build test clean cross $(addprefix cross-,$(CROSS_TARGETS))

# Native libghostty-vt only. Local development and `go test` paths use
# this; cross targets are opted into separately below so day-to-day
# builds don't pay the cost of cross-compiling every supported triple.
$(STAMP):
	cmake -B $(BUILD_DIR) -DCMAKE_BUILD_TYPE=Release
	cmake --build $(BUILD_DIR)
	@touch $(STAMP)

build: $(STAMP)
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) go build ./...

test: $(STAMP)
	PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) DYLD_LIBRARY_PATH=$(DYLD_LIBRARY_PATH) LD_LIBRARY_PATH=$(LD_LIBRARY_PATH) go test ./...

# cross builds all cross-compilation targets. Mostly useful as a local
# convenience; CI runs each cross-<target> as a separate matrix step.
cross: $(addprefix cross-,$(CROSS_TARGETS))

# cross-<target> cross-compiles the Go package for the given target
# using zig cc, and ensures libghostty-vt has been built for that same
# target via CMake. Each target gets its own stamp so reconfiguring for
# a different cross target doesn't invalidate the others.
#
# Note on escaping: the body below is processed by `$(call)` and then
# `$(eval)`, which means literal `$` characters need `$$$$` (one `$$`
# survives `call`, the second `$$` survives `eval`, leaving `$` for the
# recipe's final shell expansion). The CGO_LDFLAGS_ALLOW value contains
# a literal `$` end-anchor for the regex, which is why we need it.
define CROSS_RULE
$(BUILD_DIR)/.ghostty-built-$(1):
	cmake -B $(BUILD_DIR) -DCMAKE_BUILD_TYPE=Release \
		-DGHOSTTY_VT_CROSS_TARGETS="$(1)=$$($(1)_ZIG)"
	cmake --build $(BUILD_DIR) --target zig_build_lib_vt_$(1)
	@touch $$@

cross-$(1): $(BUILD_DIR)/.ghostty-built-$(1)
	CGO_ENABLED=1 \
	CC="zig cc -target $$($(1)_ZIG)" \
	CXX="zig c++ -target $$($(1)_ZIG)" \
	GOOS=$$($(1)_GOOS) \
	GOARCH=$$($(1)_GOARCH) \
	PKG_CONFIG_PATH="$(CURDIR)/$(BUILD_DIR)/ghostty-$(1)/share/pkgconfig" \
	CGO_LDFLAGS_ALLOW='.*[.]lib$$$$' \
	go build . ./sys/...
endef

$(foreach t,$(CROSS_TARGETS),$(eval $(call CROSS_RULE,$(t))))

clean:
	rm -rf $(BUILD_DIR)

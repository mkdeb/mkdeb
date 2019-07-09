# -*- Makefile -*-

PREFIX ?= /usr/local

GO ?= go
GOLINT ?= golint

GOOS ?= $(shell $(GO) env GOOS)
GOARCH ?= $(shell $(GO) env GOARCH)

ifeq ($(shell uname -s),Darwin)
TAR ?= gtar
else
TAR ?= tar
endif

GIT_HOOKS := $(patsubst misc/git-hooks/%,.git/hooks/%,$(wildcard misc/git-hooks/*))

BIN_LIST := $(patsubst cmd/%,%,$(wildcard cmd/*))
PKG_LIST = $(call uniq,$(dir $(wildcard */*.go)))

DIST_DIR ?= dist

DIST_ARCH = \
	darwin/386 \
	darwin/amd64 \
	freebsd/386 \
	freebsd/amd64 \
	linux/386 \
	linux/amd64 \
	linux/arm \
	linux/arm64 \
	netbsd/386 \
	netbsd/amd64 \
	openbsd/386 \
	openbsd/amd64 \
	windows/386 \
	windows/amd64

DIST_VERSION = $(shell sed -e 's/^var version = "\(.*\)"$$/\1/g' -e t -e d cmd/mkdeb/version.go)

tput = $(shell tty 1>/dev/null 2>&1 && tput $1)
print_error = (echo "$(call tput,setaf 1)Error:$(call tput,sgr0) $1" && false)
print_step = echo "$(call tput,setaf 4)***$(call tput,sgr0) $1"
uniq = $(if $1,$(firstword $1) $(call uniq,$(filter-out $(firstword $1),$1)))

all: build

clean:
	@$(call print_step,"Cleaning files...")
	@rm -rf bin/ dist/

build: build-bin

build-bin:
	@$(call print_step,"Building binaries for $(GOOS)/$(GOARCH)...")
	@for bin in $(BIN_LIST); do \
		$(GO) build -ldflags "-s -w" -mod vendor -tags "$(TAGS)" -i -o bin/$$bin -v ./cmd/$$bin || \
			$(call print_error,"failed to build $$bin for $(GOOS)/$(GOARCH)"); \
	done

test: test-bin

test-bin:
	@$(call print_step,"Testing packages...")
	@$(GO) test -cover -tags "$(TAGS)" -v $(PKG_LIST:%=./%)

install: install-bin

install-bin: build-bin
	@$(call print_step,Installing binaries...)
	@install -d -m 0755 $(PREFIX)/bin && install -m 0755 bin/* $(PREFIX)/bin/

lint: lint-bin

lint-bin:
	@$(call print_step,"Linting binaries and packages...")
	@$(GOLINT) $(BIN_LIST:%=./cmd/%) $(PKG_LIST:%=./%)

release: source
	@for arch in $(DIST_ARCH); do \
		os=$${arch%/*}; arch=$${arch#*/}; \
		$(MAKE) GOOS=$$os GOARCH=$$arch build && ( \
			test $$os = windows && \
			zip -jq $(DIST_DIR)/mkdeb_$(DIST_VERSION)_$${os}_$${arch}.zip \
				bin/* LICENSE README.md || \
			$(TAR) -czf $(DIST_DIR)/mkdeb_$(DIST_VERSION)_$${os}_$${arch}.tar.gz --transform "flags=r;s|.*\/||" \
				bin/* LICENSE README.md \
		) || exit 1; \
	done

source:
	@$(call print_step,"Building source archive...")
	@install -d -m 0755 $(DIST_DIR) && tar -czf $(DIST_DIR)/mkdeb_$(DIST_VERSION).tar.gz \
		--exclude=.git --exclude=.vscode --exclude=bin --exclude=dist .

# Always install missing Git hooks
git-hooks: $(GIT_HOOKS)

.git/hooks/%:
	@$(call print_step,"Installing $* Git hook...")
	@(install -d -m 0755 .git/hooks && cd .git/hooks && ln -s ../../misc/git-hooks/$(@F) .)

-include git-hooks

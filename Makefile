# -*- Makefile -*-

PREFIX ?= /usr/local

GO ?= go
GOLINT ?= golint

GIT_HOOKS := $(patsubst misc/git-hooks/%,.git/hooks/%,$(wildcard misc/git-hooks/*))

BIN_LIST := $(patsubst cmd/%,%,$(wildcard cmd/*))
PKG_LIST = $(call uniq,$(dir $(wildcard */*.go)))

tput = $(shell tty 1>/dev/null 2>&1 && tput $1)
print_error = (echo "$(call tput,setaf 1)Error:$(call tput,sgr0) $1" && false)
print_step = echo "$(call tput,setaf 4)***$(call tput,sgr0) $1"
uniq = $(if $1,$(firstword $1) $(call uniq,$(filter-out $(firstword $1),$1)))

all: build

clean:
	@$(call print_step,"Cleaning files...")
	@rm -rf bin/

build: build-bin

build-bin:
	@$(call print_step,"Building binaries...")
	@for bin in $(BIN_LIST); do \
		$(GO) build -i -ldflags "-s -w" -o bin/$$bin -v ./cmd/$$bin || $(call print_error,"failed to build $$bin"); \
	done

test: test-bin

test-bin:
	@$(call print_step,"Testing packages...")
	@for pkg in $(PKG_LIST); do \
		$(GO) test -cover -v ./$$pkg; \
	done

install: install-bin

install-bin: build-bin
	@$(call print_step,Installing binaries...)
	@install -d -m 0755 $(PREFIX)/bin && install -m 0755 bin/* $(PREFIX)/bin/

lint: lint-bin

lint-bin:
	@$(call print_step,"Linting binaries and packages...")
	@$(GOLINT) $(BIN_LIST:%=./cmd/%) $(PKG_LIST:%=./%)

# Always install missing Git hooks
git-hooks: $(GIT_HOOKS)

.git/hooks/%:
	@$(call print_step,"Installing $* Git hook...")
	@install -m 0755 misc/git-hooks/$* .git/hooks/$*

-include git-hooks

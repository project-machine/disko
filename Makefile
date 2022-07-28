HASH = \#
VERSION := $(shell x=$$(git describe --tags) && echo $${x$(HASH)v} || echo unknown)
VERSION_SUFFIX := $(shell [ -z "$$(git status --porcelain --untracked-files=no)" ] || echo -dirty)
VERSION_FULL := $(VERSION)$(VERSION_SUFFIX)
LDFLAGS := "${ldflags:+$ldflags }-X main.version=${ver}${suff}"
BUILD_FLAGS := -ldflags "-X main.version=$(VERSION_FULL)"
ENV_ROOT := $(shell [ "$$(id -u)" = "0" ] && echo env || echo sudo )

GOLANGCI_VER = v1.43.0
GOLANGCI = ./tools/golangci-lint-$(GOLANGCI_VER)

CMDS := demo/demo ptimg/ptimg

GO_FILES := $(wildcard *.go)
ALL_GO_FILES := $(wildcard *.go */*.go)

all: build check

build: .build $(CMDS)

.build: $(GO_FILES)
	go build ./...
	@touch $@

demo/demo: $(wildcard demo/*.go) $(GO_FILES)
	cd $(dir $@) && go build $(BUILD_FLAGS) ./...

ptimg/ptimg: $(wildcard ptimg/*.go) $(GO_FILES)
	cd $(dir $@) && go build $(BUILD_FLAGS) ./...

check: lint gofmt

gofmt: .gofmt

.gofmt: $(ALL_GO_FILES)
	o=$$(gofmt -l -w .) && [ -z "$$o" ] || { echo "gofmt made changes: $$o"; exit 1; }
	@touch $@


golangci-lint: $(GOLANGCI)

$(GOLANGCI):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b $(dir $@) $(GOLANGCI_VER) || { rm -f $(dir $@)/golangci-lint; exit 1; }
	mv $(dir $@)/golangci-lint $@

lint: .lint

.lint: $(ALL_GO_FILES) $(GOLANGCI) .golangci.yml
	$(GOLANGCI) run ./...
	@touch $@

test:
	go test -v -race -coverprofile=coverage.txt ./...

test-all:
	$(ENV_ROOT) DISKO_INTEGRATION=$${DISKO_INTEGRATION:-run} "GOCACHE=$$(go env GOCACHE)" "GOENV=$$(go env GOENV)" go test -v -coverprofile=coverage-all.tmp -count=1 ./...
	@cp coverage-all.tmp coverage-all.txt && rm -f coverage-all.tmp # dance around to not be root-owned

coverage.html: test
	go tool cover -html=coverage.txt -o $@

coverage-all.html: test-all
	go tool cover -html=coverage-all.txt -o $@

debug:
	@echo VERSION=$(VERSION)
	@echo VERSION_FULL=$(VERSION_FULL)
	@echo CMDS=$(CMDS)

clean:
	rm -f $(CMDS) coverage*.txt coverage*.html .lint .build

.PHONY: debug check test test-all gofmt clean all lint build golangci-lint

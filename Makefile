NAME := task-burner
CMDNAME := tabn
VERSION := v0.1.0
REVISION := $(shell git rev-parse --short HEAD)
GOVERSION := $(go version)

SRCS := $(shell find . -type f -name '*.go')
LDFLAGS := -ldflags="-s -w -X \"main.version=$(VERSION)\" -X \"main.revision=$(REVISION)\" -X \"main.goversion=$(GOVERSION)\" "
DIST_DIRS := find * -type d -exec

################################################################################
# Dependency And Build/Install
################################################################################

.PHONY: tidy
tidy: $(SRCS)
	export GO111MODULE=on; go mod tidy

.PHONY: verify
verify: $(SRCS)
	export GO111MODULE=on; go mod verify

.PHONY: ensure
ensure: $(SRCS) verify tidy
	export GO111MODULE=on; go mod download

.PHONY: build
build: $(SRCS) ensure
	export GO111MODULE=on; go build -i -o $(CMDNAME) $(LDFLAGS) ./...

.PHONY: install
install: $(SRCS)
	export GO111MODULE=on; go install
	mv $(GOPATH)/bin/$(NAME) $(GOPATH)/bin/$(CMDNAME)

################################################################################
# Test And Lint
################################################################################

.PHONY: test
test: $(SRCS)
	go test ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: coverage
coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

################################################################################
# Distribution
################################################################################

.PHONY: cross-build
cross-build: ensure
	for os in darwin linux windows; do \
		for arch in amd64 386; do \
			export GO111MODULE=on; GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build $(LDFLAGS) -o dist/$$os-$$arch/$(CMDNAME); \
		done; \
	done

.PHONY: dist
dist: cross-build
	cd dist && \
	$(DIST_DIRS) cp ../LICENSE {} \; && \
	$(DIST_DIRS) cp ../README.md {} \; && \
	$(DIST_DIRS) tar -zcf $(CMDNAME)-$(VERSION)-{}.tar.gz {} \; && \
	$(DIST_DIRS) zip -r $(CMDNAME)-$(VERSION)-{}.zip {} \; && \
	cd ..

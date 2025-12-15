BIN="bin"

define newline


endef

ifeq (, $(shell which golangci-lint))
$(warning "Could not find golangci-lint in PATH, run: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.7.1"$(newline))
endif

LDFLAGS :=
ifdef MDDIFF_VERSION
	LDFLAGS += -ldflags "-X mddiff/cmd.AppVersion=$(MDDIFF_VERSION)"
endif

.PHONY: all build fmt lint test clean

default: all

all: fmt test build

build:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN)/mddiff-linux-amd64 .
	GOOS=linux GOARCH=arm go build $(LDFLAGS) -o $(BIN)/mddiff-linux-arm .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BIN)/mddiff-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BIN)/mddiff-darwin-arm64 .

lint:
	$(info ******************** checking linting and formatting ********************)
	golangci-lint run

fmt:
	$(info ******************** fixing linting and formatting ********************)
	golangci-lint run --fix

test:
	$(info ******************** running tests ********************)
	go test -v ./...

clean:
	$(info ******************** cleaning up ********************)
	rm -rf $(BIN) mddiff
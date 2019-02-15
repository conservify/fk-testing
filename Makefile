GOARCH ?= amd64
GOOS ?= linux
GO ?= env GOOS=$(GOOS) GOARCH=$(GOARCH) go
BUILD ?= build
BUILDARCH ?= $(BUILD)/$(GOOS)-$(GOARCH)

all:
	GOOS=linux GOARCH=amd64 $(MAKE) binaries-all
	GOOS=linux GOARCH=arm $(MAKE) binaries-all
	GOOS=darwin GOARCH=amd64 $(MAKE) binaries-all

install: all

binaries-all: $(BUILDARCH)/fk-lan-sync $(BUILDARCH)/fk-log-analyzer $(BUILDARCH)/fk-data-tool $(BUILDARCH)/fk-wifi-tool

$(BUILD):
	mkdir -p $(BUILD)

$(BUILDARCH):
	mkdir -p $(BUILDARCH)

$(BUILDARCH)/fk-lan-sync: lan-sync/*.go utilities/*.go
	$(GO) build -o $(BUILDARCH)/fk-lan-sync lan-sync/*.go

$(BUILDARCH)/fk-log-analyzer: log-analyzer/*.go utilities/*.go
	$(GO) build -o $(BUILDARCH)/fk-log-analyzer log-analyzer/*.go

$(BUILDARCH)/fk-data-tool: data-tool/*.go utilities/*.go
	$(GO) build -o $(BUILDARCH)/fk-data-tool data-tool/*.go

$(BUILDARCH)/fk-wifi-tool: wifi-tool/*.go utilities/*.go
	$(GO) build -o $(BUILDARCH)/fk-wifi-tool wifi-tool/*.go

clean:
	rm -rf $(BUILD)

veryclean:

GOARCH ?= amd64
GOOS ?= linux
GO ?= env GOOS=$(GOOS) GOARCH=$(GOARCH) go
BUILD ?= build
BUILDARCH ?= $(BUILD)/$(GOOS)-$(GOARCH)

all:
	GOOS=linux GOARCH=amd64 make binaries-all
	GOOS=linux GOARCH=arm make binaries-all

install:
	GOOS=linux GOARCH=amd64 make binaries-install
	GOOS=linux GOARCH=arm make binaries-install

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

binaries-install: all
	cp $(BUILDARCH)/fk-lan-sync $(INSTALLDIR)
	cp $(BUILDARCH)/fk-log-analyzer $(INSTALLDIR)
	cp $(BUILDARCH)/fk-data-tool $(INSTALLDIR)
	cp $(BUILDARCH)/fk-wifi-tool $(INSTALLDIR)

clean:
	rm -rf $(BUILD)

veryclean:

GOARCH ?= amd64
GOOS ?= linux
GO ?= env GOOS=$(GOOS) GOARCH=$(GOARCH) go
BUILD_ROOT ?= build
BUILD ?= $(BUILD_ROOT)/$(GOOS)-$(GOARCH)

all:
	GOOS=linux GOARCH=amd64 make binaries
	GOOS=linux GOARCH=arm make binaries

binaries: $(BUILD)/fk-lan-sync $(BUILD)/fk-log-analyzer $(BUILD)/fk-data-tool $(BUILD)/fk-wifi-tool

$(BUILD_ROOT):
	mkdir -p $(BUILD_ROOT)

$(BUILD)/fk-lan-sync: lan-sync/*.go utilities/*.go
	$(GO) build -o $(BUILD)/fk-lan-sync lan-sync/*.go

$(BUILD)/fk-log-analyzer: log-analyzer/*.go utilities/*.go
	$(GO) build -o $(BUILD)/fk-log-analyzer log-analyzer/*.go

$(BUILD)/fk-data-tool: data-tool/*.go utilities/*.go
	$(GO) build -o $(BUILD)/fk-data-tool data-tool/*.go

$(BUILD)/fk-wifi-tool: wifi-tool/*.go utilities/*.go
	$(GO) build -o $(BUILD)/fk-wifi-tool wifi-tool/*.go

install: all
	cp $(BUILD)/fk-lan-sync $(INSTALLDIR)
	cp $(BUILD)/fk-log-analyzer $(INSTALLDIR)
	cp $(BUILD)/fk-data-tool $(INSTALLDIR)
	cp $(BUILD)/fk-wifi-tool $(INSTALLDIR)

clean:
	rm -rf $(BUILD_ROOT)

veryclean:

GOARCH ?= amd64
GO ?= env GOOS=linux GOARCH=$(GOARCH) go

all: build/fk-lan-sync build/fk-log-analyzer build/fk-data-tool build/fk-wifi-tool

build/fk-lan-sync: lan-sync/*.go
	$(GO) build -o build/fk-lan-sync lan-sync/*.go

build/fk-log-analyzer: log-analyzer/*.go
	$(GO) build -o build/fk-log-analyzer log-analyzer/*.go

build/fk-data-tool: data-tool/*.go
	$(GO) build -o build/fk-data-tool data-tool/*.go

build/fk-wifi-tool: wifi-tool/*.go
	$(GO) build -o build/fk-wifi-tool wifi-tool/*.go

install: all
	cp build/fk-lan-sync $(INSTALLDIR)
	cp build/fk-log-analyzer $(INSTALLDIR)
	cp build/fk-data-tool $(INSTALLDIR)
	cp build/fk-wifi-tool $(INSTALLDIR)

clean:
	rm -rf build

veryclean:

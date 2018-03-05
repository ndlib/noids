
GOCMD:=go
VERSION:=$(shell git describe --always)
PACKAGES:=$(shell go list ./... | grep -v /vendor/)
GO15VENDOREXPERIMENT=1

.PHONY: all test clean rpm

all: noids noid-tool/noid-tool

noids: $(wildcard *.go)
	go build .

noid-tool/noid-tool: $(wildcard noid-tool/*.go)
	cd noid-tool; go build .

test:
	$(GOCMD)  test -v $(PACKAGES)

clean:
	        rm -f noid-tool/noid-tool noids 

rpm: noids noid-tool/noid-tool
	               fpm -t rpm -s dir \
	               --name noids \
	                --version $(VERSION) \
	                --vendor ndlib \
	                --maintainer DLT \
	                --description "NOID daemon" \
	                --rpm-user app \
	                --rpm-group app \
			noids=/opt/noids/bin/noids \
			noid-tool/noid-tool=/opt/noids/bin/noid-tool

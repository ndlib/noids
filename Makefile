
BINARIES:=$(subst /bin/)
GOCMD:=go
VERSION:=$(shell git describe --always)
PACKAGES:=$(shell go list ./... | grep -v /vendor/)
GO15VENDOREXPERIMENT=1

.PHONY: all test clean rpm $(BINARIES)

all: noids noid-tool/noid-tool

noids: $(wildcard *.go) | ./bin
	go build .

noid-tool/noid-tool: $(wildcard noid-tool/*.go) | ./bin
	cd noid-tool; go build .

test:
	$(GOCMD)  test  -v $(PACKAGES)

clean:
	        rm -rf ./bin

./bin:
	mkdir -p ./bin

rpm: noids noid-tool/noid-tool
	        fpm -t rpm -s dir \
	               --name noids \
	                --version $(VERSION) \
	                --vendor ndlib \
	                --maintainer DLT \
	                --description "NOIDS daemon" \
	                --rpm-user app \
	                --rpm-group app \
			bin/noids=/opt/noids/bin/noids \
			bin/noid-tool=/opt/noids/bin/noid-tool


VERSION=$(shell cat VERSION)

.PHONY : all test update-version

all: noids noid-tool/noid-tool

noids: $(wildcard server/*.go)
	cd server; go build -o noids .
	mv server/noids .

noid-tool/noid-tool: $(wildcard noid-tool/*.go)
	cd noid-tool; go build .

test:
	go fmt ./...
	go test ./...

update-version:
	echo "package main\n\nconst version = \"$(VERSION)\"" > server/version.go
	sed -i .tmp -e "s/^Version:.*$$/Version: $(VERSION)/g" spec/noids.spec && rm -rf spec/noids.spec.tmp

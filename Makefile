
all: noids noid-tool/noid-tool

noids: $(wildcard server/*.go)
	cd server; go build -o noids .
	mv server/noids .

noid-tool/noid-tool: $(wildcard noid-tool/*.go)
	cd noid-tool; go build .

test:
	go test ./...

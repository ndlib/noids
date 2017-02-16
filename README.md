noids
=====

[![APACHE 2
License](http://img.shields.io/badge/APACHE2-license-blue.svg)](./LICENSE)
[![Contributing
Guidelines](http://img.shields.io/badge/CONTRIBUTING-Guidelines-blue.svg)](./CONTRIBUTING.md)
[![Go Report
Card](https://goreportcard.com/badge/github.com/ndlib/noids)](https://goreportcard.com/report/github.com/ndlib/noids)

Implements a server to provide a [NOID][] service.
It can persist its state to either the file system or a mysql database.
For compatibility with existing ids, the minting tries to follow
how the [Ruby Noid][] gem works.

[NOID]: https://wiki.ucop.edu/display/Curation/NOID
[Ruby Noid]: https://github.com/microservices/noid

# Installation

1. Install a [Go][] development environment. It is easiest to use a package manager.
On my Mac I use homebrew:

        brew install go

    Make sure the go version is >= 1.3

        go version

    Now set the "Go Path":

        mkdir ~/gocode
        export GOPATH=~/gocode

[Go]: http://golang.org/

2. Get the noid server:

        go get github.com/ndlib/noids

3. Run the server:

        mkdir ~/noid_pool
        $GOPATH/bin/noids --storage ~/noid_pool

# Using the API

Get a list of current noid counters:

    $ curl http://localhost:13001/pools
    [ ]

Add a new noid counter:

    $ curl http://localhost:13001/pools -F name=abc -F template=.seek
    {"Name":"abc","Template":".seek+0","Used":0,"Max":841,"Closed":false,"LastMint":"2013-12-03T11:37:14.271254-05:00"}

Mint some identifiers:

    $ curl http://localhost:13001/pools/abc/mint -F n=11
    ["000","012","024","036","048","05b","06d","07g","08j","09m","0bp"]

Get noid information:

    $ curl http://localhost:13001/pools/abc
    {"Name":"abc","Template":".seek+11","Used":11,"Max":841,"Closed":false,"LastMint":"2013-12-03T11:40:50.657972456-05:00"}

To help sync the minter with ids which have already been minted, use the AdvancePast route.
Calling this with an id will ensure that id will never be minted by this server.

    $ curl http://localhost:13001/pools/abc/advancePast -F id=bb1
    {"Name":"abc","Template":".seek+301","Used":301,"Max":841,"Closed":false,"LastMint":"2013-12-03T11:43:53.049369916-05:00"}

So now if we were to mint again:

    $ curl -X POST http://localhost:13001/pools/abc/mint
    ["bc3"]

# Using mysql database backend

MySql configuration can be done either using a config file or the command line.
The command line option `--mysql` with the connection information in the format
`user:password@tcp(hostname:port)/database`.

# Security and Authentication

There is none.

# Documentation

See [noid-service.md](noid-service.md).

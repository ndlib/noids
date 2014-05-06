noids
=====

Implements a server to provide a [NOID][] service.
There are currently a file system storage plugin.
It would be easy enough to also store information on minted noids in a database.
For compatibility with existing ids, the minting tries to follow
how the [Ruby Noid][] gem works.

[NOID]: https://wiki.ucop.edu/display/Curation/NOID
[Ruby Noid]: https://github.com/microservices/noid

# Installation

1. Install the [Go][] development environment. It is easiest to use a package manager.
On my Mac I use homebrew:

        brew install go

    Now set the "Go Path":

        mkdir ~/go
        export GOPATH=~/go

[Go]: http://golang.org/

2. Get the noid server:

        go get github.com/dbrower/noids

3. Run the server:

        mkdir ~/noid_pool
        $GOPATH/bin/noids --storage ~/noid_pool

# Using the API

Get a list of current noid counters:

    $ curl http://localhost:13001/pools
    [ ]

Add a new noid counter:

    $ curl -X POST 'http://localhost:13001/pools?name=abc&template=.seek'
    {"Name":"abc","Template":".seek+0","Used":0,"Max":841,"Closed":false,"LastMint":"2013-12-03T11:37:14.271254-05:00"}

Mint some identifiers:

    $ curl -X POST 'http://localhost:13001/pools/abc/mint?n=11'
    ["000","012","024","036","048","05b","06d","07g","08j","09m","0bp"]

Get noid information:

    $ curl http://localhost:13001/pools/abc
    {"Name":"abc","Template":".seek+11","Used":11,"Max":841,"Closed":false,"LastMint":"2013-12-03T11:40:50.657972456-05:00"}

To help sync the minter with ids which have already been minted, use the AdvancePast route.
Calling this with an id will ensure that id will never be minted by this server.

    $ curl -X POST 'http://localhost:13001/pools/abc/advancePast?id=bb1'
    {"Name":"abc","Template":".seek+301","Used":301,"Max":841,"Closed":false,"LastMint":"2013-12-03T11:43:53.049369916-05:00"}

So now if we were to mint again:

    $ curl -X POST 'http://localhost:13001/pools/abc/mint'
    ["bc3"]

# Using mysql database backend

Use the command line option `--mysql` to store the noid states in a MySQL database.
The format of the option is `user:password@tcp(hostname:port)/database`.


# Security and Authentication

There is none.

# TODO

* Standardize the naming: a noid _counter_ belongs to a _pool_, etc.

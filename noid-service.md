# Noid Minting Service API

The API uses HTTP as its transport.
No authentication or authorization is performed.
There are no rate limits or request quotas.

The service supports the creation and management of many _id pools_.
Each pool is completely independent from the others.
The ids in a pool are generated according to a template as described in the Noid specification.
Each pool has an associated reservoir of possible id values--and the reservoir may be finite or infinite depending on
the template.
A pool may be _open_ or _closed_.
An open pool may have more ids minted from it and a closed pool may not.
The intention is to allow a pool to be closed to new minting should a migration to
a newer pool is desired.
The opening and closing of a pool is controlled by the user, and a closed pool
may be reopened unless the reservoir for the pool is exhausted.
A pool automatically becomes closed when its reservoir is exhausted and then stays permanently closed.
For each pool, up to 1000 (arbitrary sane number, can be changed) ids may be requested at a single time.

Since the service may be used with id systems which are already in use, it is possible to
synchronize a pool to the existing ids.
In this case, the user ids which have already been minted, and those ids will never be returned by
the pool in the future.

### List Pools

`GET /pools`

Returns a JSON array of pool names.

### Create a new Pool

`POST /pools?name=string&template=string`

The name of the new pool is `name`.
It is an error to use a name which is currently in use.
The template of the generated identifiers is given by `template`, as described in the noid specification.
Returns a JSON object giving information on the new pool.

### Get pool information

`GET /pools/:poolname`

Returns the number of ids minted, the size of the pool, the state of the pool, the date of creation, and date of the most recent minting.

### Open or close a pool

`PUT /pools/:poolname/open`

`PUT /pools/:poolname/close`

These set the pool as either _open_ or _closed_.
If a pool is closed because it is empty, then these commands have no effect.
Returns a JSON object giving information on `:poolname`.

### Remove a pool

`DELETE /pools/:poolname`

This is not implemented, and there is no way to remove a pool once created.
If a pool is not desired, close it.
This call is here as a placeholder, it may or may not ever be added.

### Mint identifier

`POST /pools/:poolname/mint?n=50`

The optional parameter `n` is the number of identifiers to return. It should be between 1 and 1000.
If omitted, it is taken to be 1.
Returns a JSON array of the identifiers.
The array will have no more than the number asked for. But it may have less in
the case that the pool is closed or the reservoir is emptied.
Note: a closed pool will always return an empty array.
Also, minting from a closed pool is not an error.

### Server Statistics

`GET /stats`

This is not implemented.
Stats will return a JSON object containing any useful statistics which the server might have.
It also serves as a `PING` service to check if the server is up.
(Also `/pools` serves as a ping service).

### AdvancePast

`POST /pools/:poolname/advancePast`

AdvancePast is used to synchronize this noid service with any identifiers which may have previously
been minted.
It takes a mandatory parameter `id` which is an identifier conforming to the template for the pool.
The noid service then guarantees that the identifier passed will never be minted in the future.
A JSON object giving the new state of the pool is returned.
Note that advanced past may cross off more identifiers than just the one passed in, and this is
because the Noid Service thinks of each pool as consisting of a fixed sequence of identifiers.
In this sequence there is always a "next identifier to mint".
AdvancePast works by updating the "next identifier to mint".
In particular, the service does not remember every identifier actually minted.

For example, creating a pool with the template `.sdd` will mint identifiers from `00` to `99`, in numerical order.
Calling AdvancePast with `id=98` will cause the next identifier minted to be `99`, at which point the reservoir is
exhausted, and the pool is closed.

When there are already thousands of identifiers already minted, not every identifier needs to be passed to
AdvancedPast.
Because the identifiers are ordered in some way, it suffices to only past the "largest one in the sequence".
That phrase is in quotes because in the case of random identifiers it is not clear which is the largest.
The `noid-tool` command line utility can be used to determine this.

# Noid Tool

A separate command line tool provides some utilities for working with identifiers.

Usage:

	noid-tool info [<template list>]
	noid-tool valid <template> [<noid list>]
	noid-tool generate <template> [<number list>]

Modes:

 * info -- Display information about the given templates to stdout.

 * valid -- Output the sequence number of each id with respect to the given template.
Invalid ids will have a sequence number of -1.

 * generate -- Output the ids associated to each given sequence number.

If no ids are given on the command line, they will be taken from stdin,
with each id on its own line.

While being a helpful for general noid issues, this tool is mainly intended to
help transition an installation to using the noid server.
Since there will already have been ids minted, the noid server will need to be
advanced past any already minted identifiers.
The noid tool can take a dump of all these minted identifiers and figure out the
correct offset to pass to the noid server.
For example, since Fedora Commons stores identifiers in the file name, we can figure out
the correct offset using the following command line:

    $ cd fedora/data/objectStore
    $ find . -type f | cut -c 31- | noid-tool valid .reeddeeddk | sort -n | tail -1

This is just an example, the first two commands (`find` and `cut`) are used to extract the
noid from the file names.
The noid template used in the command (`.reeddeeddk`) should be changed to reflect whatever
format your application has been using.
The result of this will be a number (the index) next to an identifier.
To synchronize the noid server with this identifier, pass it in to the appropriate pool
on your noid server.
For example, if the identifier were `rj430b984` and the pool we wanted to sync was named
`xanadu-test` the following command would accomplish this.

    $ curl localhost:13001/pools/xanadu-test/advancePast -F id=rj430b984

If your application is minting a lot of identifiers during this process, mint some spoilers to
ensure duplicate identifiers won't be created in the time you are doing this step.

    $ curl localhost:13001/pools/xanadu-test/mint -F n=1000

Do this as many times as you feel necessary (1000 is the maximum number of identifiers which can be minted at one time).
For low throughput sites, it may not be necessary to do this spoiling at all.

# Noid Template Format

Noids implements the noid generator as specified in [https://wiki.ucop.edu/display/Curation/NOID]()
It aims to be compatible with the [ruby noid generator](https://github.com/microservices/noid).
In one respect, though, it is different: a noid template specifying random generation does not rely
on a random number generator. Instead it will always generate the same
sequence of ids---however the sequence generated will be scattered throughout the idspace and will
appear "random".

A noid template string has the following format:

```
    <slug> '.' <generator> <bins?> <template> <check?> <counter?>
```

Where

	<slug> is any sequence of characters (may be empty).
	<generator> is one of 'r', 's', 'z'
	<bins?> is an optional sequence of decimal digits
	<digits> is a sequence of 'd' and 'e' characters
	<check?> is an optional 'k' character
	<counter?> is an optional and is a '+' followed by decimal digits. See below

The `<bins>` element is optional, but can only be present if the generator is `r`

Example format strings:

	id.sd           -- produces id0, id1, id2, ..., id9
	id.zd           -- produces id0, id1, id2, ..., id9, id10, id11, ... unbounded
	id.sdd          -- produces id00, id01, id02, ..., id98, id99
	.reeddeeddek    -- 0000000000, 02870v839n, 05741r66m1, ... zs25x6438m, zw12z326k0
	.zddddk         -- 00000, 00014, 00028, 0003d, 0004j, ... unbounded
	.r500edek       -- 0000, 00kr, 015k, 01rb, ..., z8cn, z8zd, z9j7
	.sdek           -- 000, 012, 024, ..., 9w3, 9x5, 9z7
	a.rd.re         -- a.rd0, a.rd1, ..., a.rdz

We extend the template string with state information to completely describe
a noid minter as a string. The format is

	<noid template> '+' <number ids minted>

`<id count>` is a decimal integer >= 0. It is taken to be 0 if omitted.

For example:

	.zddddk+389
	.reeddeeddek+54321

These are the format strings returned in the pool information JSON object.

## Limitations

The implementation uses integers to represent noid counters.
Thus template strings which have between 2**63 and infinity possible identifiers
will not function correctly due to overflow (the maximum size is too large to represent).
For example the template `.seeeeeeeeeeeee` with 13 `e`s which has a
pool size of 29**13 = 10,260,628,712,958,602,189 identifiers is one such
problematic identifier.
Template strings which have infinite pools will only
have the first 2**63 identifiers minted, due to the overflow.
There are no checks in place for overflows or to identify templates which cause problems.
(It might even be a good idea to change the implementation to explicitly use
`int64` types rather than relying on `int` types implicitly being represented with `int64`.)

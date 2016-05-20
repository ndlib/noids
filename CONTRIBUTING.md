# Contributing to noids

## Getting Started
noids is a go service. There are several steps you must complete before you can to run the services, compile the code, and commit to the project.

### Development Dependencies
noids only depends on `go` in order to work correctly. `curl` is used for testing, a version of `curl` is distributed with OSX.

```console
brew update
brew install go
```

> Go is already configured by [DLT dotfiles](https://github.com/ndlib/dlt-dotfiles).

### Preparing the Environment
Go has built-in dependency management. It performs functions inside a directory that is set as the `$GOPATH`. Our convention for development environments is to set `$GOPATH` to `~/gocode`.

```console
mkdir ~/gocode
export GOPATH='~/gocode'
```

> If you manage your shell environment with [DLT dotfiles](https://github.com/ndlib/dlt-dotfiles) it will set up your `$GOPATH` for you.

To verify that `$GOPATH` is configured correctly try:

```console
echo $GOPATH
```

It should return `/Users/<YOUR_USERNAME>/gocode`.

### Checking out the Codebase
Once `$GOPATH` is configured use `go get` to check out the git repository and keep track of it so it can be included by other go projects.

```console
go get -u github.com/ndlib/noids
```

### Configuring git
The default configuration for git repositories created via `go get` is not set up to allow you to make commits back to the project. Before you make changes to the noids codebase you will need to reconfigure the git repository.

```console
cd $GOPATH/src/github.com/ndlib/noids
git remote set-url origin git@github.com:ndlib/noids.git
```

> This remote URL assumes you have commit access to ndlib/noids. If you are working on another repo, like a fork, use the URL provided buy github for that repo.

### Building noids
noids defines one executable `noids`. When the codebase in checked out using `go get` it is compiled and placed in `$GOPATH/bin`.

> `$GOPATH/bin` is already included in your `$PATH` if you use [DLT dotfiles](https://github.com/ndlib/dlt-dotfiles). Otherwise you will want to call it directly `$GOPATH/bin/noids` or include `$GOPATH/bin` in you `$PATH` manually e.g. `export $PATH=$GOPATH/bin:$PATH`.

#### If you have made changes locally
To recompile the code after you make changes use `go build`. It will create the `noids` executable at the root of the project directory. You will have to manually update the executable in `$GOPATH/bin`.

```console
cd $GOPATH/src/github.com/ndlib/noids/
go build
mv noids $GOPATH/bin/
```

> `go build` only updates the noid server. The Makefile will build both `noids` and `noid-tool`. Use the `make` command to evaluate the Makefile from the root directory of the project.

#### If you want to install the latest version from github
If you have already pushed your changes to github or if the project has been updated and you want to install the latest version use `go get`.

```console
go get -u github.com/ndlib/noids
```

## Running noids
There are also several setup steps in order to _run_ the noids server. For testing noids in development use the provided boostrap script.

```console
cd $GOPATH/src/github.com/ndlib/noids/ && ./boostrap.sh
```

This will create the needed directories, initialize the log file, start the server, create a “dev” noid pool, and display the log on STDOUT.

If you wish to configure `noids` manually continue with the provided instructions.

### Directory Setup
By convention we will run noids out of the `~/goapps` directory.

```console
mkdir -p ~/goapps/noids
```

> Creating a directory isn’t strictly needed for testing in development. By default noids stores everything in memory and writes its log to STDOUT.

### Starting the Application
There are two things you need to do before noids can mint new identifiers for you: run the service and seed the id pool.

To start the noids service:

```console
noids
```

It should start logging right away. There should be no available ID pools to start:

```console
curl localhost:13001/pools
```

Should return `[]`.

### Seeding a Pool
To create an ID pool:

```console
curl --data 'name=dev&template=.rddddd' localhost:13001/pools
```

Should return a hash of information e.g.

```console
{"Name":"dev","Template":".rddddd+0","Used":0,"Max":100000,"Closed":false,"LastMint":"2016-05-18T15:21:41.823473383-04:00"}
```

Now checking for pools should return `["dev"]`. Once there is at least one noid pool you can mint identifiers.
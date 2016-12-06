# The Service-Catalog Build System

This document aims to overview the build system in this repository.

## Composition

This repository contains several `Makefile`s that enable the
[GNU Make](https://www.gnu.org/software/make/manual/make.html) tool to build
the code herein.

Several artifacts are constructed from this repository (e.g. docker images,
binaries), and generally speaking there is a single `Makefile` per artifact.

## Structure

There is a top-level `Makefile` that drives high-level actions
(i.e. installing repository-wide Go dependencies) and builds arbitrary
other targets in each of the directories listed in the `DIRS` variable.

For example, typing `make build` at the top level will effectively run
`make -C ${DIR} build` for each `DIR` in `DIRS`.

### Subdirectory Makefiles

Each subdirectory that contains code that must be built into an artifact
has a `Makefile` in it. Generally, these `Makefile`s are simple because they
specify a few variables and include a large amount of common code. This code is
generally contained in `hack/Makefile.mk` and `hack/Common.mk`.

Here is a rough outline of one of these subdirectory `Makefile`s (eliding
the copyright notice):

```makefile
BIN=${BINARY_NAME}
PKG=github.com/kubernetes-incubator/service-catalog/${SUBDIRECTORY_PATH}
DOCKER=$(BIN)

# if the directory is more than 1 level below the top, these paths may
# need more '..' chars!
include ../hack/Makefile.mk
include ../hack/Common.mk
```

### Common Code

The above outline includes two files that provide common functionality to all
subdirectory `Makefile`s. See below for a description of each common
functionality:

- `Makefile.mk` - top-level variable definitions (e.g. the Go binary name,
  Go version and `GOPATH`) and macros to automate common functionality. These
  macros include but are not limited to:
  - Determining whether a `docker push` should be executed based on parameters
    in the subdirectory's `Makefile`. If so, executing the push
  - Determining whether a docker-container-based binary build should be executed based
    on parameters in the subdirectory's `Makefile`. If so, executing the
    container-based build
    in the subdirectory's `Makefile`. If so, executing the build
- `Common.mk` - common build targets (i.e. `build`) which are designed to read
   variables set by the subdirectory's `Makefile` and execute them accordingly.
   For example, if the `BIN` variable is set in a subdirectory `Makefile`,
   the `build` target will run an appropriate `go build` command. Otherwise,
   that target is a no-op

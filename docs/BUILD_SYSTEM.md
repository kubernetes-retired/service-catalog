# The Service-Catalog Build System

This document presents an overview of the build system in this repository.

## Composition

This repository contains several `Makefile`s that enable the
[GNU Make](https://www.gnu.org/software/make/manual/make.html) tool to build
the code herein.

Several components are built from this repository (catalog controller,
example brokers, ...). One or more artifacts are constructed for each component
(e.g. docker images, binaries). Generally speaking there is a single
`Makefile` per component.

## Structure

There is a top-level [Makefile](./Makefile) that drives high-level actions
(i.e. installing repository-wide Go dependencies) and recursively builds other
targets in each of the directories listed in the `DIRS` variable.

For example, typing `make build` at the top level will effectively run
`make -C ${DIR}` build for each component of the project.

The subdirectory `Makefile`s are designed to be used both by calling
`make <target>` from top level (in which case `make <target>` in all
subdirectories will be executed) but also as standalone `Makefile`s. This is
to make it possible to run `make` in each individual subdirectory, and
operate on subset of the code base.


## Subdirectory `Makefile`s

Each subdirectory that contains code that must be built into an artifact has
a `Makefile` in it. Generally, these `Makefile`s are simple because they specify
a few variables, and the remainder of the `Makefile` code is included from
[`hack/Makefile.mk`](./hack/Makefile.mk) and
[`hack/Common.mk`](./hack/Common.mk).

Below is a rough outline of one of these subdirectory `Makefile`s (eliding the
copyright notice):

```console
BIN=${BINARY_NAME}
PKG=github.com/kubernetes-incubator/service-catalog/${SUBDIRECTORY_PATH}
DOCKER=${DOCKER_IMAGE_NAME, most commonly $(BIN)}

# if the directory is more than 1 level below the top, these paths may
# need more '..' chars!
# the make executes in the current directory and include clauses use relative paths
# to include Makefile.mk and Common.mk.
include ../hack/Makefile.mk
include ../hack/Common.mk
```

## Common Code

Two shared `Makefile`s provide common functionality to all subdirectory
`Makefile`s: [`Makefile.mk`](./hack/Makefile.mk) and
[`Common.mk`](./hack.Common.mk). See below for a description of each common
file:

- [`Makefile.mk`](./hack/Makefile.mk) - top-level variable definitions (e.g.
  the Go binary name, Go version and GOPATH) and macros to automate common
  functionality. These macros include but are not limited to:
    - Determining whether a docker push should be executed based on parameters
      in the subdirectory's `Makefile`. If so, executing the push
    - Determining whether a docker-container-based binary build should be
      executed based on parameters in the subdirectory's `Makefile`. If so,
      executing the container-based build in the subdirectory's `Makefile`
    - Defines the `all` target, which loves to be the first target defined in
      the `Makefile` and is therefore defined in
      [`Makefile.mk`](./hack/Makefile.mk) which is included first
- [`Common.mk`](./hack/Common.mk) - defines all common build targets (e.g.
  `build`, `test`) except for `all`. The common build targets use the
  variables defined by the subdirectory's `Makefile` to define the make
  targets. For example, if the `BIN` variable is set in a subdirectory's
  `Makefile`, the build target will run an appropriate go build command.
  Otherwise, that target is a no-op

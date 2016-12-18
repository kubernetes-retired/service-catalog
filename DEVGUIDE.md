# Developer's Guide to Service-Catalog

## Building

To build the service-catalog you have two options:
* `make build`
* `DOCKER=1 make build`

Both will build all of the executables, into the `bin` directory. However,
the second option will do the build within a Docker container - meaning you
do not need to have all of the necessary tooling installed on your host
(such as a golang compiler or glide). Whichever option you choose, the
results should be the same.

## Testing

Currently, we only have unit testcases within this repo:
* `make test`
* `DOCKER=1 make test`

These will execute any `*_test.go` files within the source tree.

To see how well these tests cover the source code, you can use:
* `make coverage`
* `DOCKER=1 make coverage`

These will execute the tests and perform an analysis of how well they
cover all code paths. The results are put into a file called:
`coverage.html` at the root of the repo.


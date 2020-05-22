---
title: Developer Guide
layout: docwithnav
---

## Overview

Thank you for deciding to contribute to our project! 💖 We welcome contributors
from all backgrounds and experience levels. 

If you are interested in going beyond a single PR, take a look at our 
[contribution ladder](/contribute/ladder.md) and learn how to become a reviewer, 
or even a maintainer!

## Working on Issues

Github does not allow non-maintainers to assign, or be assigned to, issues.
As such non-maintainers can indicate their desire to work on (own) a particular
issue by adding a comment to it of the form:

	#dibs

However, it is a good idea to discuss the issue, and your intent to work on it,
with the other members via the [slack channel](https://kubernetes.slack.com/messages/sig-service-catalog)
to make sure there isn't some other work already going on with respect to that
issue.

When you create a pull request (PR) that completely addresses an open issue
please include a line in the initial comment that looks like:

	Closes: #1234

where `1234` is the issue number. This allows Github to [automatically
close the issue](https://help.github.com/articles/closing-issues-using-keywords/)
when the PR is merged.

Also, before you start working on your issue, please read our [Code Standards](./code-standards.md)
document.

## Prerequisites

At a minimum you will need:

* [Docker](https://www.docker.com) 17.05+ installed locally (configured with 4+ GB RAM)
* GNU Make
* [git](https://git-scm.com)

These will allow you to build and test service catalog components within a
Docker container.

If you want to deploy service catalog components built from source, you will
also need:

* A working Kubernetes cluster and `kubectl` installed in your local `PATH`,
  properly configured to access that cluster. The version of Kubernetes and
  `kubectl` must be >= 1.12. See below for instructions on how to download these
  versions of `kubectl`
* [Helm](https://helm.sh) (Tiller) installed in your Kubernetes cluster and the
  `helm` binary in your `PATH`
* To be pre-authenticated to a Docker registry (if using a remote cluster)

**Note:** It is not generally useful to run service catalog components outside
a Kubernetes cluster. As such, our build process only supports compilation of
linux/amd64 binaries suitable for execution within a Docker container.

## Workflow
We can set up the repo by following a process similar to the [dev guide for k8s]( https://github.com/kubernetes/community/blob/master/contributors/devel/development.md#1-fork-in-the-cloud)

### 1 Fork in the Cloud
1. Visit https://github.com/kubernetes-sigs/service-catalog
2. Click Fork button (top right) to establish a cloud-based fork.

### 2 Clone fork to local storage

From your shell:
```bash

# Set user to match your github profile name
user={your github profile name}

# Create your clone:
mkdir -p $working_dir
cd $working_dir
git clone https://github.com/$user/service-catalog.git
# or: git clone git@github.com:$user/service-catalog.git

cd service-catalog
git remote add upstream https://github.com/kubernetes-sigs/service-catalog.git
# or: git remote add upstream git@github.com:kubernetes-sigs/service-catalog.git

# Never push to upstream master
git remote set-url --push upstream no_push

# Confirm that your remotes make sense:
git remote -v
```

## Code Layout
This repository is organized as similarly to Kubernetes itself as the developers
have found possible (or practical). Below is a summary of the repository's
layout:

    .
    ├── bin                     # Destination for binaries compiled for linux/amd64 (untracked)
    ├── build                   # Contains build-related scripts and subdirectories containing Dockerfiles
    ├── charts                  # Helm charts for deployment
    │   ├── catalog             # Helm chart for deploying the service catalog
    │   └── ups-broker          # Helm chart for deploying the user-provided service broker
    ├── cmd                     # Contains "main" Go packages for each service catalog component binary
    │   ├── controller-manager  # The service catalog controller manager service-catalog command
    │   ├── service-catalog     # The service catalog binary, which is used to run commands
    │   ├── svcat               # The command-line interface for interacting with kubernetes service-catalog resources
    │   └── webhook             # The service catalog webhook server command
    ├── contrib                 # Contains examples, non-essential golang source, CI configurations, etc
    │   ├── build               # Dockerfiles for contrib images (example: ups-broker)
    │   ├── cmd                 # Entrypoints for contrib binaries
    │   ├── examples            # Example API resources
    │   ├── hack                # Non-build related scripts
    │   │   ├── ci              # CI configuration
    │   │   └── ...             # Rest helper bash scripts
    │   └── pkg                 # Contrib golang code
    ├── docs                    # Documentation
    ├── pkg                     # Contains all non-"main" Go packages
    ├── plugin                  # Plugins for API server
    ├── test                    # Integration and e2e tests
    ├── vendor                  # dep-managed dependencies
    ├── go.mod                  # defines projects requirements and locks dependencies
    └── go.sum                  # the expected cryptographic checksums of go.mod dependencies

## Building

First `cd` to the root of the cloned repository tree.
To build the service-catalog server components:

    $ make build

The above will build all executables and place them in the `bin` directory. This
is done within a Docker container-- meaning you do not need to have all of the
necessary tooling installed on your host (such as a golang compiler or dep).
Building outside the container is possible, but not officially supported.

To build the service-catalog client, `svcat`:

    $ make svcat

The svcat cli binary is located at `bin/svcat/svcat`.

To install `svcat` to your $GOPATH/bin directory:

    $ make svcat-install

Note, this will do the basic build of the service catalog. There are more
more [advanced build steps](#advanced-build-steps) below as well.

To deploy to Kubernetes, see the
[Deploying to Kubernetes](#deploying-to-kubernetes) section.

### Notes Concerning the Build Process/Makefile

* The Makefile assumes you're running `make` from the root of the repo.
* There are some source files that are generated during the build process.
  These are:

    * `pkg/client/*_generated`
    * `pkg/apis/servicecatalog/zz_*`
    * `pkg/apis/servicecatalog/v1beta1/zz_*`
    * `pkg/apis/servicecatalog/v1beta1/types.generated.go`
    * `pkg/openapi/openapi_generated.go`

* Running `make clean` or `make clean-generated` will roll back (via
  `git checkout --`) the state of any generated files in the repo.
* Running `make purge-generated` will _remove_ those generated files from the
  repo.
* A Docker Image called "scbuildimage" will be used. The image isn't pre-built
  and pulled from a public registry. Instead, it is built from source contained
  within the service catalog repository.
* While many people have utilities, such as editor hooks, that auto-format
  their go source files with `gofmt`, there is a Makefile target called
  `format` which can be used to do this task for you.
* `make build` will build binaries for linux/amd64 only.

## Testing

There are three types of tests: unit, integration and e2e.

### Unit Tests

The unit testcases can be run via the `test-unit` Makefile target, e.g.:

    $ make test-unit

These will execute any `*_test.go` files within the source tree.

### Integration Tests

The integration tests can be run via the `test-integration` Makefile target,
e.g.:

    $ make test-integration

The integration tests require the Kubernetes client (`kubectl`) so there is a
script called `contrib/hack/kubectl` that will run it from within a
Docker container. This avoids the need for you to download, or install it,
youself. You may find it useful to add `contrib/hack` to your `PATH`.

### E2E Tests

The e2e tests are executed against Kubernetes cluster with service-catalog deployed 
into it. The e2e testcases can be run via the `test-e2e` Makefile target:

```console
 $ make test-e2e
```

Sample test output:
```console
I0816 17:20:37.451423   75760 e2e.go:45] Starting e2e run "e39cfcd2-cbae-41fc-96ee-447095a492bd" on Ginkgo node 1
Running Suite: Service Catalog e2e suite
========================================
Random Seed: 1565968837 - Will randomize all specs
Will run 5 of 5 specs

< ... Test Output ... >

•
Ran 5 of 5 Specs in 49.761 seconds
SUCCESS! -- 5 Passed | 0 Failed | 0 Pending | 0 Skipped --- PASS: TestE2E (49.76s)
```

> **NOTE:** Docker is required for running e2e tests locally.

Under the hood, the script executes such flow:

1. Install [Helm](https://github.com/helm/helm) and the [kind](https://github.com/kubernetes-sigs/kind) tool.
2. Provision a Kubernetes cluster using the kind tool.
3. Build Service Catalog images from sources. 
4. Deploy Service Catalog into cluster. 
5. Execute [e2e tests](../test/e2e).

   If any test fails, then [cluster info](https://github.com/kubernetes/kubernetes/blob/release-1.14/pkg/kubectl/cmd/clusterinfo/clusterinfo_dump.go#L93-L96) from the namespace where the Service Catalog is installed is dumped.
    
6. Delete the Kubernetes cluster.

### Test Running Tips

The `test` Makefile target will run both the unit and integration tests, e.g.:

    $ make test

If you want to run just a subset of the unit testcases then you can
specify the source directories of the tests:

    $ TEST_DIRS="path1 path2" make test

or you can specify a regexp expression for the test name:

    $ UNIT_TESTS=TestBar* make test

a regexp expression also works for integration test names:

    $ INT_TESTS=TestIntegrateBar* make test

You can also set the log level for the tests, which is useful for
debugging using the `TEST_LOG_LEVEL` env variable. Log level 5 e.g.:

    $ TEST_LOG_LEVEL=5 make test-integration

### Test Code Coverage

To see how well these tests cover the source code, you can use:

    $ make coverage

These will execute the tests and perform an analysis of how well they
cover all code paths. The results are put into a file called:
`coverage.html` at the root of the repo.

As mentioned above, integration tests require a running Catalog API & ETCD image
and a properly configured .kubeconfig.  When developing or drilling in on a
specific test failure you may find it helpful to run Catalog in your "normal"
environment and as long as you have properly configured your KUBECONFIG
environment variable you can run integration tests much more quickly with a
couple of commands:

    $ make build-integration
    $ ./integration.test -test.v -v 5 -logtostderr -test.run  TestPollServiceInstanceLastOperationSuccess/async_provisioning_with_error_on_second_poll

The first command ensures the test integration executable is up-to-date.  The
second command runs one specific test case with verbose logging and can be
re-run over and over without having to wait for the start and stop of API and
ETCD.  This example will execute the test case "async provisioning with error on
second poll" within the integration test
TestPollServiceInstanceLastOperationSuccess.

### Golden Files
The svcat tests rely on "[golden files](https://medium.com/@povilasve/go-advanced-tips-tricks-a872503ac859#a196)",
a pattern used in the Go standard library, for testing command output. The expected
output is stored in a file in the testdata directory, `cmd/svcat/testdata`, and
and then the test's output is compared against the "golden output" stored
in that file. It helps avoid putting hard coded strings in the tests themselves.

You do not edit the golden files by hand. When you need to update the golden
files, run `make test-update-goldenfiles` or `go test ./cmd/svcat/... -update`,
and the golden files are updated automatically with the results of the test run.

For new tests, first you need to manually create the empty golden file into the destination
directory specified in your test, e.g. `touch cmd/svcat/testdata/mygoldenfile.txt`
before updating the golden files. This only manages the contents of the golden files,
but doesn't create or delete them.

Keep in mind that golden files help catch errors when the output unexpectedly changes.
It's up to you to judge when you should run the tests with -update,
and to diff the changes in the golden file to ensure that the new output is correct.

### Counterfeiter
Certain tests use fakes generated with [Counterfeiter](http://github.com/maxbrunsfeld/counterfeiter). If you add a method
to an interface (such as SvcatClient in pkg/svcat/service-catalog) you may need to regenerate the fake. You can install
Counterfeiter by running `go get github.com/maxbrunsfeld/counterfeiter`.
Then regenerate the fake with `counterfeiter ./pkg/svcat/service-catalog SvcatClient` and manually paste the boilerplate
copyright comment into the generated file.

## FeatureGates
Feature gates are a set of key=value pairs that describe experimental features
and can be turned on or off by specifying the value when launching the Service
Catalog executable (typically done in the Helm chart).  A new feature gate
should be created when introducing new features that may break existing
functionality or introduce instability.  See [FeatureGates](feature-gates.md)
for more details.

When adding a FeatureGate to Helm charts, define the variable
`fooEnabled` with the `false` value in [values.yaml](https://github.com/kubernetes-sigs/service-catalog/blob/master/charts/catalog/values.yaml).  
In the [Webhook Server](https://github.com/kubernetes-sigs/service-catalog/blob/master/charts/catalog/templates/webhook-deployment.yaml) and [Controller](https://github.com/kubernetes-sigs/service-catalog/blob/master/charts/catalog/templates/controller-manager-deployment.yaml)
templates, add the new FeatureGate:
{% raw %}
```yaml
    - --feature-gates
    - Foo={{.Values.fooEnabled}}
```
{% endraw %}

When the feature has had enough testing and the community agrees to change the
default to true, update [features.go](https://github.com/kubernetes-sigs/service-catalog/blob/master/pkg/features/features.go) and `values.yaml` changing the default for
feature foo to `true`. And lastly update the appropriate information in the
[FeatureGates doc](feature-gates.md).

## Documentation

Our documentation site is located at [svc-cat.io](https://svc-cat.io). The content files are located
in the `docs/` directory, and the website framework in `docsite/`.

To preview your changes, run `make docs-preview` and then open `http://localhost:4000` in
your web browser. When you create a pull request, you can preview documentation changes by
clicking on the `deploy/netlify` build check in your PR.

## Making a Contribution

Once you have compiled and tested your code locally, make a Pull
Request. Create a branch on your local repo with a short descriptive
name of the work you are doing. Make a commit with the work in it, and
push it up to your remote fork on github. Come back to the code tab of
the repository, and there should be a box suggesting to make a Pull
Request.

Pull requests are expected to have a few things before asking people to review the PR:

* [Build the code](#building) with `make build` (for server-side changes) or `make svcat` (for cli changes).
* [Run the tests](#testing) with `make test`.
* Run the build checks with `make verify`. This helps catch compilation errors
and code formatting/linting problems.
* Added new tests or updated existing tests to verify your changes. If this is a svcat related change,
you may need to [update the golden files](#golden-files).
* Any associated documentation changes. You can preview documentation changes by
clicking on the `deploy/netlify` build check on your pull request.

After you create a PR, relevant CI tests need to complete successfully.
If you are not a Kubernetes, contact the repository maintainers specified in the CODEOWNERS file to review 
your PR and add the [ok-to-test](https://prow.k8s.io/command-help#ok_to_test) label to your PR to trigger all tests.

If a test fails, check the reason by clicking the Details button next to the given job on your PR. 
Make the required changes and the tests rerun. If you want to run a specific test, 
add the /test {test-name} or /retest {test-name} comment to your PR. To rerun all failed tests, add the /retest comment.

You can use the [Prow /cc command](https://prow.k8s.io/command-help#cc)
to request reviews from the maintainers of the project. This works even
if you do not have status in the service-catalog project.

## Advanced Build Steps

You can build the service catalog executables into Docker images yourself. By
default, image names are `quay.io/kubernetes-service-catalog/<component>`. Since
most contributors who hack on service catalog components will wish to produce
custom-built images, but will be unable to push to this location, it can be
overridden through use of the `REGISTRY` environment variable.

Examples of service-catalog image names:

| `REGISTRY` | Fully Qualified Image Name | Notes |
|----------|----------------------------|-------|
| Unset; default | `quay.io/kubernetes-service-catalog/service-catalog` | You probably don't have permissions to push to here |
| Dockerhub username + trailing slash, e.g. `krancour/` | `krancour/service-catalog` | Missing hostname == Dockerhub |
| Dockerhub username + slash + some prefix, e.g. `krancour/sc-` | `krancour/sc-service-catalog` | The prefix is useful for disambiguating similarly names images within a single namespace. |
| 192.168.99.102:5000/ | `192.168.99.102:5000/service-catalog` | A local registry |

With `REGISTRY` set appropriately:

    $ make images push

This will build Docker images for all service catalog components. The images are
also pushed to the registry specified by the `REGISTRY` environment variable, so
they can be accessed by your Kubernetes cluster.

The images are tagged with the current Git commit SHA:

    $ docker images

### svcat targets
These are targets for the service-catalog client, `svcat`:

* `make svcat-all` builds all supported client platforms (darwin, linux, windows).
* `make svcat-for-X` builds a specific platform.
* `make svcat` builds for the current dev's platform.
* `make svcat-publish` compiles everything and uploads the binaries.

The same tags are used for both client and server. The cli uses the format that
always includes a tag, so that it's clear which release you are "closest" to,
e.g. v1.2.3 for official releases and v1.2.3-2-gabc123 for untagged commits.

### Deploying Releases

* Merge to master - A docker image for the server is pushed to [quay.io/kubernetes-service-catalog/service-catalog](http://quay.io/kubernetes-service-catalog/service-catalog),
  tagged with the abbreviated commit hash. Nothing is deployed for the client, `svcat`.
* Tag a commit on master with vX.Y.Z - A docker image for the server is pushed,
  tagged with the version, e.g. vX.Y.Z. The client binaries are published to
  https://download.svcat.sh/cli/latest/OS/ARCH/svcat and https://download.svcat.sh/cli/VERSION/OS/ARCH/svcat.

The idea behind "latest" link is that we can provide a permanent link to the most recent stable release of `svcat`.
If someone wants to install a unreleased version, they must build it locally.

----

## Deploying to Kubernetes

Use the [`catalog` chart](../charts/catalog) to deploy the service
catalog into your cluster.  The easiest way to get started is to deploy into a
cluster you regularly use and are familiar with.

If you have recently merged changes that haven't yet made it into a
release, you probably want to deploy the canary images. Always use the
canary images when testing local changes.

For more information see the
[installation instructions](./install.md). The last two lines of
the following `helm install` example show the canary images being
installed with the other standard installation options.

From the root of this repository:

```
helm install charts/catalog \
    --name catalog --namespace catalog \
    --set image=quay.io/kubernetes-service-catalog/service-catalog:canary
```

### Deploy local canary
For your convenience, you can use the following script quickly rebuild, push and
deploy the canary image. There are a few assumptions about your environment and
configuration in the script. If the assumptions do
not match your needs, we suggest copying the contents of that script and using
it as a starting off point for your own custom deployment script.

```console
# The registry defaults to DockerHub with the same user name as the current user
# Examples: quay.io/myuser/service-catalog/, another-user/
$ export REGISTRY="myuser/"
$ ./contrib/hack/deploy-local-canary.sh
```

## Dependency Management

This section is intended to show a way for managing `vendor/` tree dependencies using go modules.

### Theory of operation

The `go.mod` file describes dependencies using two directives:

* `require` directives list the preferred version of dependencies (this is auto-updated by go tooling to the maximum preferred version of the module)
* `replace` directives pin to specific tags or commits

### Adding or updating a dependency

The most common things people need to do with dependencies are adding and updating them.
These operations are handled the same way:

For the sake of examples, consider that we have discovered a wonderful Go
library at `example.com/go/foo`.

Step 1: Ensure there is go code in place that references the packages you want to use.
```go
import "example.com/go/foo"
// ...
frob.DoStuff()
```

Step 2: Determine what version of the dependency you want to use, and add that version to the go.mod file:

```sh
contrib/hack/pin-dependency.sh example.com/go/foo v1.0.4
```

This fetches the dependency, resolves the specified sha or tag, and adds two entries to the `go.mod` file:

```
require (
    example.com/go/foo v1.0.4
    ...
)

replace (
    example.com/go/foo => example.com/go/foo v1.0.4
    ...
)
```

The `require` directive indicates our module requires `example.com/go/foo` >= `v1.0.4`.
If our module was included as a dependency in a build with other modules that also required `example.com/go/foo`,
the maximum required version would be selected (unless the main module in that build pinned to a lower version).

The `replace` directive pins us to the desired version when running go commands e.g. build svcat binary.

Step 3: Rebuild the `vendor` directory:

```sh
contrib/hack/update-vendor.sh
```

Step 4: Check if the new dependency was added correctly:

```sh
contrib/hack/lint-dependencies.sh
```

### Removing a dependency

This happens almost for free.  If you edit Service Catalog code and remove the last
use of a given dependency, you only need to run `contrib/hack/update-vendor.sh`, and the
tooling will figure out that you don't need that dependency anymore and remove it,
along with any unused transitive dependencies.

### Reviewing and approving dependency changes

Particular attention to detail should be exercised when reviewing and approving
PRs that add/remove/update dependencies. Importing a new dependency should bring
a certain degree of value as there is a maintenance overhead for maintaining
dependencies into the future.

When importing a new dependency, be sure to keep an eye out for the following:
- Is the dependency maintained?
- Does the dependency bring value to the project? Could this be done without adding a new dependency?
- Is the target dependency the original source, or a fork?
- Is there already a dependency in the project that does something similar?
  
>**NOTE:** Always check if there is a tagged release we can vendor instead of a random hash


## Demo walkthrough

Check out the [walkthrough](./walkthrough.md) to get started with
installation and a self-guided demo.

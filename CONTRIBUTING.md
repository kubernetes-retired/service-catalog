# Contributing to service-catalog

This document should concisely express the project development status,
methodology, and contribution process.  As the community makes progress, we
should keep this document in sync with reality.

# Claiming an Issue

We use the [standard k8's labels](https://github.com/kubernetes/community/blob/master/contributors/devel/help-wanted.md)
[good first issue][good-first-issue] and [help wanted][help-wanted]
to indicate issues that are ideal for new contributors.

Once you have found an issue that you'd like to work on, comment on it with
"#dibs", or "I would like to work on this". If someone else said that they would
like to work on it, but there's no open PR and it's been more than 2 weeks,
comment with "@kubernetes-incubator/maintainers-service-catalog Is it okay if I
take this?" and a maintainer will help out.

## Submitting a Pull Request (PR)

The following outlines the general rules we follow:

- All PRs must have the appropriate documentation changes made within the
same PR. Note, not all PRs will necessarily require a documentation change
but if it does please include it in the same PR so the PR is complete and
not just a partial solution.
- All PRs must have the appropriate testcases. For example, bug fixes should
include tests that demonstrates the issue w/o your fix. New features should
include as many testcases, within reason, to cover any variants of use of the
feature.
- All PRs must have appropriate documentation. New features should be
  described, and an example of use provided.
- PR authors will need to have CLA on-file with the Linux Foundation before
the PR will be merged.
See Kubernete's [contributing guidelines](https://github.com/kubernetes/kubernetes/blob/master/CONTRIBUTING.md) for more information.

See our [reviewing PRs](REVIEWING.md) documentation for how your PR will
be reviewed.

## Development status

We're currently collecting use-cases and requirements for our [v1 milestone](./docs/v1).

## Methodology

Each milestone will have a directory within the [`docs`](./docs) directory of
this project.   We will keep a complete record of all supported use-cases and
major designs for each milestone.

## Contributing to a release

If you would like to propose or change a use-case, open a pull request to the
project, adding or altering a file within the `docs` directory.

We'll update this space as we begin developing code with relevant dev
information.

[help-wanted]: https://github.com/kubernetes-incubator/service-catalog/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22
[good-first-issue]: https://github.com/kubernetes-incubator/service-catalog/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22+

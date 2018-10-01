# Reviewing and Merging Pull Requests

This document is a guideline for how core contributors should review and merge
pull requests (PRs). It is intended to outline the lightweight process that
we’ll use for now. It’s assumed that we’ll operate on good faith for now in
situations for which process is not specified.

PRs are automatically merged after the following criteria are met:

1. It has the `lgtm` label applied. This label is automatically removed when
    the commits in the PR are modified. It can be added with `/lgtm` and removed
    with `/lgtm cancel`.
1. It has the `approved` label applied. This label is "sticky" and remains
    even after subsequent changes are made to the commits in the PR. It can be
    added with `/approve` and removed with `/approve cancel`.
1. The CI checks are all passing.

## Holds

If a PR should not be merged in its current state,
even once it has the `lgtm` and `approved` labels from others, mark that PR with
`do-not-merge/hold` label using the `/hold` command.

This label should only be used by a reviewer when that person believes there
is a fundamental problem with the PR. The reviewer should summarize that problem
in the PR comments and a longer discussion may be required.

We expect this label to be used infrequently.

# Alerts

You can join the [SIG Service Catalog Alerts](https://groups.google.com/forum/#!forum/kubernetes-sig-service-catalog-alerts)
mailing list to receive notifications when there are problems with master or release builds.

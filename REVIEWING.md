# Reviewing and Merging Pull Requests

This document is a guideline for how core contributors should review and merge
pull requests (PRs). It is intended to outline the lightweight process that
we’ll use for now. It’s assumed that we’ll operate on good faith for now in
situations for which process is not specified.

PRs may only be merged after the following criteria are met:

1. It has been approved by 2 different reviewers, each from a different
  organization and different from the author's organization
1. It has all appropriate corresponding documentation and test cases

We use the the 
[Github Pull Request Reviews](https://help.github.com/articles/about-pull-request-reviews/)
system to approve PRs. 

## LGTMs

When a reviewer believes that a PR is ready to merge, they should submit their PR review
as "Approve" to the pull request.

If they still have comments that they would like the submitter to address,
and don't believe it's ready to review, they should submit their comments as part of the Github 
PR review system, and then submit the review as "Request changes".

Details on how to use the pull request reviews system can be found
[here](https://help.github.com/articles/about-pull-request-reviews/).

## Vetoing

If a reviewer decides that a PR should not be merged in its current state,
even if it has 2 approvals from others, they should mark that PR with
`do-not-merge` label.

This label should only be used by a reviewer when that person believes there
is a fundamental problem with the PR, and it needs to be discussed in more detail before it
proceeds.

The reviewer should summarize that problem in a PR comment and organize a longer discussion
if necessary.

We expect the `do-not-merge` label to be used infrequently for the purpose of preventing a PR
merge.

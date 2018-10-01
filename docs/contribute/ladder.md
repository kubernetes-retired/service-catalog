---
title: Contribution Ladder
layout: docwithnav
---

There are many ways to contribute to the Service Catalog project. Here are the
various levels of involvement and how to get there!

In general we try to follow the [standard Kubernetes roles](https://github.com/kubernetes/community/blob/master/community-membership.md).

## User
Try out Service Catalog, either on development cluster like Minikube or in the 
cloud. Let us know how it goes!

  * Open an issue for any gaps in the documentation that you encounter.
  * File a bug when unexpected things happen.
  * Take a look at some of our open proposals and provide feedback.
  * Answer questions of other users in the community.

## Member
Project members are people who use Service Catalog and are interested in contributing
futher.

* Assign issues to yourself.
* Verify that a pull request is safe to run tests against with the `/ok-to-test`
  command.
* Review a pull request and apply the `/lgtm` command that signals that it's
  ready for final approval.

Once you feel ready, ask another member if they would be willing to be a sponsor
for your [membership request](https://github.com/kubernetes/org/issues/new?template=membership.md&title=REQUEST%3A%20New%20membership%20for%20%3Cyour-GH-handle%3E).
The best people to ask are those who have reviewed your pull requests in the past.
When you request membership, make sure to ask for the `kubernetes-incubator` 
organization.
 
✅ If you are already a member of the github.com/kubernetes organization, you can
immediately apply to become a member of github.com/kubernetes-incubator where
the Service Catalog project lives.

❓If you aren't sure if you are ready, or need help finding sponsors, reach out
to the current [chairs][chairs].

## Reviewer
After you have contributed to an area of Service Catalog for a while, you may
ready help review pull requests either in just that area, or for the entire 
repository.

Reviewers have an entry in the [OWNERS](https://github.com/kubernetes/community/blob/master/contributors/guide/owners.md)
file in the repository indicating that they are a good candidate to be 
automatically assigned or suggested as a reviewer for that area of code 
(or the entire repository).

❓If you are willing to review PRs in a particular area, submit a PR adding
your GitHub name to the OWNERS file in the sub-directory of that area in a 
`reviewers` section.

## Maintainer
Kubernetes traditionally calls this role "approver".

Maintainers have an approver entry in the [OWNERS](https://github.com/kubernetes/community/blob/master/contributors/guide/owners.md)
file in the repository indicating that they are an experienced reviewer and
contributor in a particular area. They are responsible for the final review of 
a pull request, and signing off that it is ready to merge.

* Review a pull request and apply the final `/approve` command that signals 
  that the pull request is ready to merge.
* An approver may decide to apply both the `/lgtm` and `/approve` commands. This
  sometimes is a good choice for small non-controversial pull requests where
  there aren't other people who should be consulted first.
* Our repository is configured to require both the `lgtm` and `approved` labels
  before merging. 

❓If you feel that you are ready to become a maintainer, reach out to a [chair][chairs]
and they will help sponsor you. Maintainers are added by a vote amongst the chairs.

See our [charter][charter] for a full list of responsibilities and the voting process.

## Chair
Chairs are maintainers who also perform extra project management and 
administrative work for the project such as:

* Facilitating the SIG meetings and recording them.
* Grooming the backlog.
* Finding issues that are good for beginners and ensuring that issues are
  appropriately explained and labeled.
* Representing the SIG at Kubernetes community standup meetings.

❓If you are interested in becoming a chair, reach out to a [chair][chairs]
and they will help sponsor you. Chairs are added by a vote amongst the chairs.

See our [charter][charter] for a full list of responsibilities and the voting process.

[charter]: https://github.com/kubernetes/community/blob/master/sig-service-catalog/charter.md
[chairs]: https://github.com/kubernetes/community/blob/master/sig-service-catalog/README.md#chairs

# Kubernetes repository infrastructure

This repository contains repository infrastructure tools for use in
`kubernetes` and `kubernetes-incubator` repositories.  Examples:

- Boilerplate verification
- Gofmt verification
- Golang build infrastructure

---

## Using this repository

The `repo-infra` repository is designed to be used via
[git subtree](http://git.kernel.org/cgit/git/git.git/plain/contrib/subtree/git-subtree.txt)
and placed in the top level of your project.

`repo-infra expects to be placed into the top level of your project under repo-
`infra:

```
repository-root/  # eg, service-catalog
  repo-infra/
    ...
```

### Adding a `repo-infra` subtree

To add `repo-infra` to your repository, use the following commands from the root
directory of **your** repository.

First, add a git remote for the `repo-infra` repository:

```
$ git remote add repo-infra git://github.com/kubernetes/repo-infra
```

This is not strictly necessary, but reduces the typing required for subsequent
commands.

Next, use `git subtree add` to create a new subtree in the `repo-infra`
directory within your project:

```
$ git subtree add -P repo-infra repo-infra master --squash
```

After this command, you will have:

1.  A `repo-infra` directory in your project containing the content of **this**
    project
2.  2 new commits in the active branch:
  1.  A commit that squashes the git history of the `repo-infra` project
  2.  A merge commit whose ancestors are:
    1.  The `HEAD` of the branch prior to when you ran `git subtree add`
    2.  The commit containing the squashed `repo-infra` commits


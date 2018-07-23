 <!--  Thanks for sending a pull request!  Here are some tips for you:
1. If this is your first time, read our contributor guidelines https://github.com/kubernetes-incubator/service-catalog/blob/master/CONTRIBUTING.md
2. If the PR is unfinished, see how to mark it: https://git.k8s.io/community/contributors/guide/pull-requests.md#marking-unfinished-pull-requests
3. See our merge process https://github.com/kubernetes-incubator/service-catalog/blob/master/REVIEWING.md
-->

This PR is a 
 - [ ] Feature Implementation
 - [ ] Bug Fix
 - [ ] Documentation

**What this PR does / why we need it**:

**Which issue(s) this PR fixes** 
<!-- *(optional, in `fixes #<issue number>(, fixes #<issue_number>, ...)` format, will close the issue(s) when PR gets merged)*: -->
Fixes #

Please leave this checklist in the PR comment so that maintainers can ensure a good PR.

Merge Checklist:
 - [ ] New feature 
   - [ ] Tests
   - [ ] Documentation
 - [ ] SVCat CLI flag
 - [ ] Server Flag for config
   - [ ] Chart changes
   - [ ] removing a flag by marking deprecated and hiding to avoid
         breaking the chart release and existing clients who provide a
         flag that will get an error when they try to update

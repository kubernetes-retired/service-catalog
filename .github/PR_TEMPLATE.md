
This PR is a 
 - [ ] Feature Implementation
 - [ ] Bug Fix
 - [ ] Documentation

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

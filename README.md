[![CircleCI](https://circleci.com/gh/giantswarm/ci-cleaner/tree/master.svg?style=svg&circle-token=200804d99fdd5ee97482012f23d4470b62f8e34c)](https://circleci.com/gh/giantswarm/ci-cleaner/tree/master)
[![Docker Repository on Quay](https://quay.io/repository/giantswarm/ci-cleaner/status "Docker Repository on Quay")](https://quay.io/repository/giantswarm/ci-cleaner)

# CI Cleaner

Cleans up cloud provider resources created during tests in CI (continuous integration).



### AWS

In AWS, this cleans up:

- CloudFormation stacks
  - that are older than 90 minutes
  - matching certain name prefixes (`cluster-ci-`, `host-peer-ci-`, `e2e-`)
- S3 buckets
  - that are older than 90 minutes
  - matching certain name criteria (please see source code)

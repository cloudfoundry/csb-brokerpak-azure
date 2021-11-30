## Upgrade tests

These tests perform the following flow:
- Push version A of this brokerpak (with broker)
- Create some resources
- Push version B of this brokerpak (with broker)
- Check that the resources are still accessible
- Create some more resources

Two flags may be specified to control the versions used:
- `-releasedBuildDir` specifies a directory with a built brokerpak at version A
- `-developmentBuildDir` specifies a directory with a built brokerpak at version B

Note that when running Ginkgo, these flags should go after a `--` to distinguish
them from Ginkgo flags, for example:
```
ginkgo -v -- -releasedBuildDir ... -developmentBuildDir ...
```
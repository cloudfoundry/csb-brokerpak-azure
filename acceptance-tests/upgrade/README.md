## Upgrade tests

These tests perform the following flow:
- Push version A of this brokerpak (with broker)
- Create some resources
- Push version B of this brokerpak (with broker)
- Check that the resources are still accessible
- Create some more resources

### Usage
By default, version A will be the highest brokerpak version on GitHub and
version B will be whatever has been built locally with `make build`.

### Specifying version A
To specify version A, use the `-from-version` flag:
```
ginkgo -v --label-filter <label> -- -from-version 1.10.0
```
Note the `--` to separate Ginkgo flags from test flags.

### Advanced flags
Two other flags may be specified to control the versions used:
- `-releasedBuildDir` specifies a directory with a built brokerpak at version A
- `-developmentBuildDir` specifies a directory with a built brokerpak at version B.
  By default it is `../..` which is the root of this repo.
- `intermediateBuildDirs` comma separated locations of intermediate versions of built broker and brokerpak

```
ginkgo -v -- -releasedBuildDir ... -developmentBuildDir ...
```
Note the `--` to separate Ginkgo flags from test flags.

### Helper commands

To list brokerpak versions available:
```
go run -C ../helpers/brokerpaks/versions .
```

To prepare a brokerpak without running a test:
```
go run -C ../helpers/brokerpaks/prepare . -version 1.10.0 -dir /tmp/versions
```
Both flags are optional.
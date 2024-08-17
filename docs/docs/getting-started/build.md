---
sidebar_position: 5
---

# Build from source

The Pouch codebase builds with [Bazel](https://bazel.build). We recommend
using [Bazelisk](https://github.com/bazelbuild/bazelisk) to remain up to date with Bazel version updates inside Pouch.

## Build everything

Running build from the repository root builds a binary for the host platform, and multi-platform OCI images.

```shell
bazelisk build
```
To build binaries just for your platform
```shell
bazelisk build //server:bin
```

## Run all tests

All the tests can be run from root.

```shell
bazelisk test //...
```
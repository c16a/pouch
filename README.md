# Pouch

A really tiny KV store.

## Building from source

The repository build system outputs binaries for both the server and the CLI. Additionally, OCI containers are also
build for multiple architectures.

## Running tests

All tests can be run from the repository root.

```shell
bazelisk test //...
```

```shell
bazelisk build //...
```
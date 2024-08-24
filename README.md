# Pouch

A really tiny KV store.

## Building from source

The repository build system outputs binaries for both the server and the CLI. Additionally, OCI containers are also
built for multiple architectures.

```shell
bazelisk build //...
```

## Running tests

All tests can be run from the repository root.

```shell
bazelisk test //...
```
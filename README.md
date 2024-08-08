# Pouch

A really tiny KV store.

## Building from source

Both the `server` and the `cli` can be built from the repository root. The binaries can be found at `target/release`
or `target/debug` depending on the profile.

## Running tests

All tests can be run from the repository root.
```shell
cargo test
```

```shell
# For debug builds
cargo build

# For release builds
cargo build --release
```
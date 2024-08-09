# Pouch

A really tiny KV store.

## Crates

| Crate  | Version                                                                                                    |
|--------|------------------------------------------------------------------------------------------------------------|
| Server | [![pouch-server](https://img.shields.io/crates/v/pouch-server.svg)](https://crates.io/crates/pouch-server) |
| CLI    | [![pouch-server](https://img.shields.io/crates/v/pouch-cli.svg)](https://crates.io/crates/pouch-cli)       |
| SDK    | [![pouch-server](https://img.shields.io/crates/v/pouch-sdk.svg)](https://crates.io/crates/pouch-sdk)       |

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
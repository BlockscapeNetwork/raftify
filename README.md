# Raftify

[![CircleCI](https://circleci.com/gh/BlockscapeLab/raftify/tree/master.svg?style=shield)](https://circleci.com/gh/BlockscapeLab/raftify/tree/master)
[![codecov](https://codecov.io/gh/BlockscapeLab/raftify/branch/master/graph/badge.svg)](https://codecov.io/gh/BlockscapeLab/raftify)
[![Go Report Card](https://goreportcard.com/badge/github.com/blockscapelab/raftify)](https://goreportcard.com/report/github.com/blockscapelab/raftify)
[![License](https://img.shields.io/github/license/cosmos/cosmos-sdk.svg)](https://github.com/cosmos/cosmos-sdk/blob/master/LICENSE)

> :warning: This project has not yet had a security audit or stress test and is therefore not ready for use in production! Use at your own risk!

_Raftify_ is a Go implementation of the Raft leader election algorithm without the Raft log and enables the creation of a self-managing cluster of nodes by transforming an application into a Raft node. It is meant to be a **more cost-efficient** small-scale alternative to running a validator cluster with a separate full-fledged Raft consensus layer.

It is designed to be directly embedded into an application and provide a direct way of communicating between individual nodes, omitting the overhead caused by replicating a log.
Raftify was built with one particular use case in mind: **running a self-managing cluster of Cosmos validators**.

## Requirements

- Golang 1.14+

## Configuration Reference

The configuration is to be provided in a `raftify.json` file and must be located in the working directory specified in the second parameter of the `InitNode` method.

> :information_source: For Gaia, the working directory is `~/.gaiad/config/` by default.

| Key         | Value    | Description                                                                                                                                                                                                           |
|:------------|:---------|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `id`          | string   | **(Mandatory)** The node's identifier.</br>Must be **unique**.                                                                                                                                                         |
| `max_nodes`   | int      | **(Mandatory)** The self-imposed limit of nodes to be run in the cluster.</br>Must be greater than 0 and must _never_ be exceeded. |
| `expect`      | int      | **(Mandatory)** The number of nodes expected to be online in order to bootstrap the cluster and start the leader election. Once the expected number of nodes is online, all cluster members will be started simultaneously.</br>Must be 1 or higher and must _never_ exceed the self-imposed `max_nodes` limit.</br>:warning: Please use `expect = 1` for single-node setups only. If you plan on running more than one node, set the `expect` value to the final cluster size on **ALL** nodes. |
| `encrypt`     | string   | _(Optional)_ The hex representation of the secret key used to encrypt messages.</br>The value must be either 16, 24, or 32 bytes to select AES-128, AES-192, or AES-256.</br>[**Use this tool to generate a key.**](https://www.browserling.com/tools/random-bytes) |
| `performance` | int      | _(Optional)_ The modifier used to multiply the maximum and minimum timeout and ticker settings. Higher values increase leader stability and reduce bandwidth and CPU but also increase the time needed to recover from a leader failure.</br>Must be 1 or higher. Defaults to 1 which is also the maximum performance setting. |
| `log_level`   | string   | _(Optional)_ The minimum log level for console log messages.</br>Can be DEBUG, INFO, WARN, ERR. Defaults to `WARN`.                                                                                                    |
| `bind_addr`   | string   | _(Optional)_ The address to bind the node application to.</br>Defaults to `0.0.0.0`.                                                                                                                                                        |
| `bind_port`   | string   | _(Optional)_ The port to bind the node application to.</br>Defaults to `7946`.                                                                                                                                                              |
| `peer_list`   | []string | _(Optional)_ The list of IP addresses of all cluster members (optionally including the address of the local node). It is used to determine the quorum in a non-bootstrapped cluster.</br>For example, if your peerlist has `n = 3` nodes then `math.Floor((n/2)+1) = 2` nodes will need to be up and running to bootstrap the cluster.</br>Addresses must be provided in the `host:port` format.</br>Must not be empty if more than one node is expected. |

### Example Configuration

```json
{
    "id": "My-Unique-Name",
    "max_nodes": 3,
    "expect": 3,
    "encrypt": "8ba4770b00f703fcc9e7d94f857db0e76fd53178d3d55c3e600a9f0fda9a75ad",
    "performance": 1,
    "log_level": "WARN",
    "bind_addr": "192.168.0.25",
    "bind_port": 3000,
    "peer_list": [
        "192.168.0.25:3000",
        "192.168.0.26:3000",
        "192.168.0.27:3000"
    ]
}
```

## Usage

### Step 1

Get the latest version of [raftify-cosmos-sdk](https://github.com/BlockscapeLab/raftify-cosmos-sdk).

### Step 2

Get [Gaia](https://github.com/cosmos/gaia) and check out the latest version.

Once you have checked out the latest version, open up the `go.mod` file and add the following line at the very bottom:

```go
replace github.com/cosmos/cosmos-sdk => github.com/BlockscapeLab/raftify-cosmos-sdk v0.37.9-R1
```

All that is left to do now is to build Gaia. For more information on how to do that, check out Gaia's [Makefile](https://github.com/cosmos/gaia/blob/master/Makefile).

## Testing

Use

```go
make unit-tests
```

to run unit tests, and

```go
make integration-tests
```

to run integration tests.

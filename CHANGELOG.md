# Changelog

## Unreleased

### Breaking Changes

* `InitNode` is now blocking. This means that on startup up a node, it will only unblock once the cluster is properly bootstrapped, that is if the expected number of nodes have formed a cluster.

### Bugfixes

* Fixed a bug that prevented nodes with `expect = 1` from becoming the cluster leader if there were other peers listed in the peerlist of the raftify.json file.

### Testing

* Added more unit tests for more stable code coverage

## v0.1.0

* First release

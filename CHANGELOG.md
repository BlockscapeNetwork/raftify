# Changelog

## Unreleased

### Breaking Changes

* `InitNode` is now blocking. This means that on startup up a node, it will only unblock once the cluster is properly bootstrapped, that is if the expected number of nodes have formed a cluster.

### General Changes

* Bump memberlist to `v0.2.2`
* Raftify is now able to distinguish between intended and crash- or timeout-related leave events. This allows it to immediately adjust the quorum for intended leave events instead of having to wait for the dead nodes to be kicked out of the cluster.

### Bugfixes

* Fixed a bug that prevented nodes with `expect = 1` from becoming the cluster leader if there were other peers listed in the peerlist of the raftify.json file.

## v0.1.0

* First release

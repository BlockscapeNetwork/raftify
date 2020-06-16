# Changelog

## Unreleased

### Breaking Changes

* `InitNode` is now blocking. Nodes will only unblock once the cluster is properly bootstrapped, that is if the expected number of nodes have formed a cluster.

### General Changes

* Bump to memberlist `v0.2.2`

### Bugfixes

* Fixed a bug that prevented nodes with `expect = 1` from becoming the cluster leader if there were other peers listed in the peerlist of the raftify.json file.

## v0.1.0

* First release

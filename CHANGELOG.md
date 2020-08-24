# Changelog

## Unreleased

---

### General Changes

* Added custom handling for intended leave events in order to distinguish them better from crash-related ones

## v0.2.0

---

### Breaking Changes

* `InitNode` is now blocking. This means that on startup up a node, it will only unblock once the cluster is properly bootstrapped, that is if the expected number of nodes have formed a cluster

### Bugfixes

* Merge pull request #5: Fixed a bug that prevented nodes with `expect = 1` from becoming the cluster leader if there were other peers listed in the peerlist of the raftify.json file
* Merge pull request #9: Fixed a bug that prevented the prevote quorum from being adjusted to the new cluster size
* Merge pull request #10: Fixed a bug that caused a node to get stuck after a rejoin during operation

### General Changes

* Bump to memberlist `v0.2.2`
* Merge pull request #11: Added version info on startup

### Testing

* Merge pull request #7: Added more unit tests for more stable code coverage

## v0.1.0

---

* First release

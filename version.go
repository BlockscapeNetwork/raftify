package raftify

import (
	"fmt"
	"runtime"
)

// VersionInfo defines version information about Raftify.
type VersionInfo struct {
	Name      string
	Version   string
	GoVersion string
}

// GetVersionInfo returns version information.
func (v VersionInfo) GetVersionInfo() VersionInfo {
	return VersionInfo{
		Name:      "Raftify",
		Version:   "v0.3.0",
		GoVersion: fmt.Sprintf("%v %v/%v", runtime.Version(), runtime.GOOS, runtime.GOARCH),
	}
}

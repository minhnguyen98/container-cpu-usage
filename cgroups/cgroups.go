//go:build linux
// +build linux

package cgroups

import (
	"sync"

	"golang.org/x/sys/unix"
)

const (
	cgroupMountPoint  = "/sys/fs/cgroup"
	procCGroupPath    = "/proc/self/cgroup"
	procMountInfoPath = "/proc/self/mountinfo"
)

type cgroup interface {
	cpuQuota() (float64, error)
	cpuUsage() (uint64, error)
	effectiveCPUs() (int, error)
}

var (
	checkMode sync.Once
	isUnified bool
)

func newCGroup() (cgroup, error) {
	if isUnifiedMode() {
		return newCGroupV2()
	}

	return newCGroupV1()
}

// mode returns the cgroups mode running on the host
func isUnifiedMode() bool {
	checkMode.Do(func() {
		var st unix.Statfs_t
		err := unix.Statfs(cgroupMountPoint, &st)
		if err != nil {
			return
		}
		isUnified = st.Type == unix.CGROUP2_SUPER_MAGIC
	})

	return isUnified
}

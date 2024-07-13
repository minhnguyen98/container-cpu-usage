//go:build linux
// +build linux

package cgroups

import (
	"os"
	"path"
	"strings"
)

type cgroupv1 struct {
	cgroups map[string]string
}

func newCGroupV1() (*cgroupv1, error) {
	subsystems, err := parseCGroupSubsystems(procCGroupPath)
	if err != nil {
		return nil, err
	}

	mountInfos, err := getMountInfos(procMountInfoPath, fsTypeFilter("cgroup"))
	if err != nil {
		return nil, err
	}

	cgroups := make(map[string]string)

	for _, mountInfo := range mountInfos {
		for _, opt := range mountInfo.SuperOptions {
			_, exists := subsystems[opt]
			if !exists {
				continue
			}

			cgroups[opt] = path.Join(cgroupMountPoint, opt)
		}
	}

	return &cgroupv1{
		cgroups: cgroups,
	}, nil
}

// cpuQuota returns the CPU quota applied with the CPU cgroup controller.
// It is a result of `cpu.cfs_quota_us / cpu.cfs_period_us`.
func (cg *cgroupv1) cpuQuota() (float64, error) {
	cpuCGroupPath, exists := cg.cgroups["cpu"]
	if !exists {
		return -1, nil
	}

	cpuQuotaUs, err := readInt(path.Join(cpuCGroupPath, "cpu.cfs_quota_us"))
	if defined := cpuQuotaUs > 0; err != nil || !defined {
		return -1, err
	}

	cpuPeriodUs, err := readInt(path.Join(cpuCGroupPath, "cpu.cfs_period_us"))
	if defined := cpuPeriodUs > 0; err != nil || !defined {
		return -1, err
	}

	return float64(cpuQuotaUs) / float64(cpuPeriodUs), nil
}

// cpuUsage returns the CPU usage applied with the CPU cgroup controller.
// cpuacct.usage gives the CPU time (in nanoseconds) obtained by this group
// which is essentially the CPU time obtained by all the tasks
// in the system.
// https://man7.org/linux/man-pages/man7/cgroups.7.html
// https://www.kernel.org/doc/Documentation/cgroup-v1/cpuacct.txt
func (cg *cgroupv1) cpuUsage() (uint64, error) {
	cpuCGroupPath, exists := cg.cgroups["cpuacct"]
	if !exists {
		return 0, nil
	}

	data, err := os.ReadFile(path.Join(cpuCGroupPath, "cpuacct.usage"))
	if err != nil {
		return 0, err
	}

	return parseUint(strings.TrimSpace(string(data)))
}

// effectiveCPUs returns the CPU effective for cgroup controller.
// cpuset.cpus is a list of the physical numbers of the CPUs on which
// processes in that cpuset are allowed to execute.
// https://man7.org/linux/man-pages/man7/cpuset.7.html
// https://www.kernel.org/doc/Documentation/admin-guide/cgroup-v1/cpusets.rst
func (cg *cgroupv1) effectiveCPUs() (int, error) {
	cpuCGroupPath, exists := cg.cgroups["cpuset"]
	if !exists {
		return 0, nil
	}

	data, err := readFirstLine(path.Join(cpuCGroupPath, "cpuset.cpus"))
	if err != nil {
		return 0, err
	}

	cpus, err := parseUints(data)
	if err != nil {
		return 0, err
	}

	return len(cpus), nil
}

//go:build linux
// +build linux

package cgroups

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type cgroupv2 struct {
	path    string
	cgroups map[string]string
}

func newCGroupV2() (*cgroupv2, error) {
	subsystems, err := parseCGroupSubsystems(procCGroupPath)
	if err != nil {
		return nil, err
	}

	// For the cgroups version 2 hierarchy, this field contains the value 0
	var v2subsys *Subsystem
	for _, subsys := range subsystems {
		if subsys.ID == 0 {
			v2subsys = subsys
			break
		}
	}

	if v2subsys == nil {
		return nil, errors.New("cgroupv2 subsystem is nil")
	}
	cgroups := make(map[string]string)
	path := filepath.Join(cgroupMountPoint, v2subsys.Name)
	if err := readKVStatsFile(path, "cpu.stat", cgroups); err != nil {
		return nil, err
	}

	return &cgroupv2{
		path:    path,
		cgroups: cgroups,
	}, nil
}

func readKVStatsFile(path string, file string, out map[string]string) error {
	f, err := os.Open(filepath.Join(path, file))
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) != 2 {
			return errors.New("cgroups: parsing file with invalid format failed")
		}

		out[fields[0]] = fields[1]
	}
	return s.Err()
}

// cpuQuota returns the CPU quota applied with the CPU cgroup2 controller.
// It is a result of reading cpu quota and period from cpu.max file.
// It will return `cpu.max / cpu.period`.
func (cg *cgroupv2) cpuQuota() (float64, error) {
	cpuMaxFile, err := os.Open(path.Join(cg.path, "cpu.max"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return -1, nil
		}
		return -1, err
	}
	defer cpuMaxFile.Close()

	scanner := bufio.NewScanner(cpuMaxFile)
	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		// fields has format: [$MAX $PERIOD]
		if len(fields) != 2 {
			return -1, fmt.Errorf("invalid cgroupV2 with line: %s", scanner.Text())
		}

		if fields[0] == "max" {
			return -1, nil
		}

		max, err := strconv.Atoi(fields[0])
		if err != nil {
			return -1, err
		}

		period, err := strconv.Atoi(fields[1])
		if err != nil {
			return -1, err
		}

		if period == 0 {
			return -1, errors.New("zero value for period is not allowed")
		}

		return float64(max) / float64(period), nil
	}

	if err := scanner.Err(); err != nil {
		return -1, err
	}

	return 0, errors.New("fail to parse cpu quota cgroupV2")
}

// cpuUsage returns the CPU total usage for cgroup2 controller.
// It is a result of reading cpu usage from cpu.stat file with field usage_usec.
// https://www.kernel.org/doc/Documentation/cgroup-v2.txt
func (cg *cgroupv2) cpuUsage() (uint64, error) {
	// Example of cpu.stat format:
	// usage_usec 20905476302
	// user_usec 20039242823
	// system_usec 866233479
	// All time durations are in microseconds.
	usec, err := parseUint(cg.cgroups["usage_usec"])
	if err != nil {
		return 0, err
	}

	return usec * uint64(time.Microsecond), nil
}

// effectiveCPUs returns the CPU effective for cgroup2 controller in cpuset.
// cpuset.cpus is a list of the physical numbers of the CPUs on which
// processes in that cpuset are allowed to execute.
// https://man7.org/linux/man-pages/man7/cpuset.7.html
// https://www.kernel.org/doc/Documentation/admin-guide/cgroup-v1/cpusets.rst
func (cg *cgroupv2) effectiveCPUs() (int, error) {
	data, err := readFirstLine(path.Join(cg.path, "cpuset.cpus.effective"))
	if err != nil {
		return 0, err
	}

	cpus, err := parseUints(data)
	if err != nil {
		return 0, err
	}

	return len(cpus), nil
}

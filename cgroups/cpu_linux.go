//go:build linux
// +build linux

package cgroups

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	clockTicksPerSecond = 100
)

var (
	preSystem uint64
	preTotal  uint64
	limit     float64
	cores     uint64
	noCgroup  bool
	initOnce  sync.Once
)

func CollectCPUUsage() (float64, float64) {
	initializeOnce()

	if noCgroup {
		return 0, 0
	}

	total, err := cpuUsage()
	if err != nil {
		return 0, 0
	}

	system, onlineCPUs, err := systemCPUUsage()
	if err != nil {
		return 0, 0
	}

	usage, percent := calculateCPUUsage(total, system, onlineCPUs)

	preSystem = system
	preTotal = total

	return usage, percent
}

func calculateCPUUsage(total, system, onlineCPUs uint64) (float64, float64) {
	var (
		cpuUsage    = 0.0
		cpuPercent  = 0.0
		cpuDelta    = float64(total) - float64(preTotal)
		systemDelta = float64(system) - float64(preSystem)
		cpuCores    = float64(onlineCPUs)
	)

	if cpuCores == 0.0 {
		cpuCores = float64(cores)
	}

	if systemDelta > 0 && cpuDelta > 0 {
		cpuUsage = (cpuDelta / systemDelta) * cpuCores
		if limit > 0 {
			cpuPercent = cpuDelta * float64(cores) * 100 / (systemDelta * limit)
		}
	}

	return cpuUsage, cpuPercent
}

func initializeOnce() {
	initOnce.Do(func() {
		defer func() {
			if p := recover(); p != nil {
				noCgroup = true
			}
		}()

		if err := initialize(); err != nil {
			noCgroup = true
		}
	})
}

func initialize() error {
	cpus, err := effectiveCpus()
	if err != nil {
		return err
	}

	cores = uint64(cpus)
	limit = float64(cpus)
	quota, err := cpuQuota()
	if err == nil && quota > 0 {
		if quota < limit {
			limit = quota
		}
	}

	preSystem, _, err = systemCPUUsage()
	if err != nil {
		return err
	}

	preTotal, err = cpuUsage()

	return err
}

func cpuQuota() (float64, error) {
	cg, err := newCGroup()
	if err != nil {
		return -1, err
	}

	return cg.cpuQuota()
}

func cpuUsage() (uint64, error) {
	cg, err := newCGroup()
	if err != nil {
		return 0, err
	}

	return cg.cpuUsage()
}

func effectiveCpus() (int, error) {
	cg, err := newCGroup()
	if err != nil {
		return 0, err
	}

	return cg.effectiveCPUs()
}

// systemCPUUsage returns the host system's cpu usage in
// nanoseconds and number of online CPUs. An error is returned
// if the format of the underlying file does not match.
//
// Uses /proc/stat defined by POSIX. Looks for the cpu
// statistics line and then sums up the first seven fields
// provided. See `man 5 proc` for details on specific field
// information.
// https://github.com/moby/moby/blob/master/daemon/stats_unix.go#L321
func systemCPUUsage() (cpuUsage uint64, cpuNum uint64, err error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return 0, 0, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 4 || line[:3] != "cpu" {
			break // Assume all cpu* records are at the front, like glibc https://github.com/bminor/glibc/blob/5d00c201b9a2da768a79ea8d5311f257871c0b43/sysdeps/unix/sysv/linux/getsysstats.c#L108-L135
		}
		if line[3] == ' ' {
			parts := strings.Fields(line)
			if len(parts) < 8 {
				return 0, 0, fmt.Errorf("invalid number of cpu fields")
			}
			var totalClockTicks uint64
			for _, i := range parts[1:8] {
				v, err := strconv.ParseUint(i, 10, 64)
				if err != nil {
					return 0, 0, fmt.Errorf("unable to convert value %s to int: %w", i, err)
				}
				totalClockTicks += v
			}
			cpuUsage = (totalClockTicks * uint64(time.Second)) / clockTicksPerSecond
		}
		if '0' <= line[3] && line[3] <= '9' {
			cpuNum++
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, fmt.Errorf("error scanning '/proc/stat' file: %w", err)
	}

	return
}

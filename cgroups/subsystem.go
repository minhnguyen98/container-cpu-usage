//go:build linux
// +build linux

package cgroups

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// the data structure for entities in `/proc/$PID/cgroup`.
// See also cgroups(7) for more information.
// https://man7.org/linux/man-pages/man7/cgroups.7.html
type Subsystem struct {
	ID         int
	Subsystems []string
	Name       string
}

// parseCGroupSubsysFromLine returns a new *Subsystem by parsing a string in
// the format of `/proc/pid/cgroup`
// https://man7.org/linux/man-pages/man7/cgroups.7.html
func parseCGroupSubsysFromLine(line string) (*Subsystem, error) {
	fields := strings.SplitN(line, ":", 3)

	if len(fields) < 3 {
		return nil, fmt.Errorf("invalid cgroup entry: %q", line)
	}

	id, err := strconv.Atoi(fields[0])
	if err != nil {
		return nil, err
	}

	cgroup := &Subsystem{
		ID:         id,
		Subsystems: strings.Split(fields[1], ","),
		Name:       fields[2],
	}

	return cgroup, nil
}

// parseCGroupSubsystems parses procPathCGroup (`/proc/pid/cgroup`)
func parseCGroupSubsystems(path string) (map[string]*Subsystem, error) {
	cgroupFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer cgroupFile.Close()

	scanner := bufio.NewScanner(cgroupFile)
	subsystems := make(map[string]*Subsystem)

	for scanner.Scan() {
		cgroup, err := parseCGroupSubsysFromLine(scanner.Text())
		if err != nil {
			return nil, err
		}
		for _, subsys := range cgroup.Subsystems {
			subsystems[subsys] = cgroup
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return subsystems, nil
}

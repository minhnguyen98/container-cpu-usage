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

// FilterFunc used to filter out mountinfo entries
// skip: true if the entry should be skipped
type filterFunc func(*MountInfo) (skip bool)

// A MountInfo is a type that describes the details, options
// for each mount, parsed from /proc/self/mountinfo.
// The fields described in each entry of /proc/self/mountinfo
// is described in the following man page.
// http://man7.org/linux/man-pages/man5/proc.5.html
// /proc/pid/mountinfo
type MountInfo struct {
	MountID        int
	ParentID       int
	MajorMinorVer  string
	Root           string
	MountPoint     string
	Options        []string
	OptionalFields []string
	FSType         string
	MountSource    string
	SuperOptions   []string
}

func parseMountInfoString(mountString string) (*MountInfo, error) {
	var err error
	fields := strings.Split(mountString, " ")
	numFields := len(fields)
	if numFields < 10 {
		// should be at least 10 fields
		return nil, fmt.Errorf("parsing '%s' failed: not enough fields (%d)", mountString, numFields)
	}

	if fields[numFields-4] != "-" {
		return nil, fmt.Errorf("couldn't find separator in expected field: %s", fields[numFields-4])
	}

	mountID, err := strconv.Atoi(fields[0])
	if err != nil {
		return nil, err
	}
	parentID, err := strconv.Atoi(fields[1])
	if err != nil {
		return nil, err
	}

	mount := &MountInfo{
		MountID:        mountID,
		ParentID:       parentID,
		MajorMinorVer:  fields[2],
		Root:           fields[3],
		MountPoint:     fields[4],
		Options:        strings.Split(fields[5], ","),
		OptionalFields: nil,
		FSType:         fields[numFields-3],
		MountSource:    fields[numFields-2],
		SuperOptions:   strings.Split(fields[numFields-1], ","),
	}

	// Has optional fields, which is a space separated list of values.
	// Example: shared:2 master:7
	if fields[6] != "" {
		mount.OptionalFields = fields[6 : numFields-4]
	}
	return mount, nil
}

// getMounts retrieves mountinfo information from `/proc/self/mountinfo`.
func getMountInfos(path string, filter filterFunc) ([]*MountInfo, error) {
	mountInfoFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer mountInfoFile.Close()

	mounts := []*MountInfo{}
	scanner := bufio.NewScanner(mountInfoFile)
	for scanner.Scan() {
		mountString := scanner.Text()
		parsedMounts, err := parseMountInfoString(mountString)
		if err != nil {
			return nil, err
		}

		if filter != nil {
			skip := filter(parsedMounts)
			if skip {
				continue
			}
		}

		mounts = append(mounts, parsedMounts)
	}

	err = scanner.Err()
	return mounts, err
}

// fsTypeFilter returns all entries that match provided fstype(s).
func fsTypeFilter(fstype ...string) filterFunc {
	return func(m *MountInfo) bool {
		for _, t := range fstype {
			if m.FSType == t {
				return false // don't skip
			}
		}
		return true // skip
	}
}

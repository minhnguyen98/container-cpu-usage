//go:build linux
// +build linux

package cgroups

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCGroupSubsysFromLine(t *testing.T) {
	testTable := []struct {
		name           string
		line           string
		expectedSubsys *Subsystem
	}{
		{
			name: "single-subsys",
			line: "1:cpu:/",
			expectedSubsys: &Subsystem{
				ID:         1,
				Subsystems: []string{"cpu"},
				Name:       "/",
			},
		},
		{
			name: "multi-subsys",
			line: "8:cpu,cpuacct,cpuset:/docker/1234567890abcdef",
			expectedSubsys: &Subsystem{
				ID:         8,
				Subsystems: []string{"cpu", "cpuacct", "cpuset"},
				Name:       "/docker/1234567890abcdef",
			},
		},
		{
			name: "multi-subsys",
			line: "12:cpu,cpuacct:/system.slice/containerd.service/kubepods-besteffort-podb41662f7_b03a_4c65_8ef9_6e4e55c3cf27.slice:cri-containerd:1753b7cbbf62734d812936961224d5bc0cf8f45214e0d5cdd1a781a053e7c48f",
			expectedSubsys: &Subsystem{
				ID:         12,
				Subsystems: []string{"cpu", "cpuacct"},
				Name:       "/system.slice/containerd.service/kubepods-besteffort-podb41662f7_b03a_4c65_8ef9_6e4e55c3cf27.slice:cri-containerd:1753b7cbbf62734d812936961224d5bc0cf8f45214e0d5cdd1a781a053e7c48f",
			},
		},
	}

	for _, tt := range testTable {
		subsys, err := parseCGroupSubsysFromLine(tt.line)
		assert.Equal(t, tt.expectedSubsys, subsys, tt.name)
		assert.NoError(t, err, tt.name)
	}
}

func TestParseCGroupSubsysFromLineErr(t *testing.T) {
	lines := []string{
		"1:cpu",
		"not-a-number:cpu:/",
	}
	_, parseError := strconv.Atoi("not-a-number")

	testTable := []struct {
		name          string
		line          string
		expectedError error
	}{
		{
			name:          "fewer-fields",
			line:          lines[0],
			expectedError: fmt.Errorf("invalid cgroup entry: %q", lines[0]),
		},
		{
			name:          "illegal-id",
			line:          lines[1],
			expectedError: parseError,
		},
	}

	for _, tt := range testTable {
		subsys, err := parseCGroupSubsysFromLine(tt.line)
		assert.Nil(t, subsys, tt.name)
		assert.Equal(t, tt.expectedError, err, tt.name)
	}
}

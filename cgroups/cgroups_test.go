//go:build linux
// +build linux

package cgroups

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCgroups(t *testing.T) {
	// test cgroup legacy(v1) & hybrid
	if !isUnifiedMode() {
		cg, err := newCGroupV1()
		assert.NoError(t, err)
		_, err = cg.effectiveCPUs()
		assert.NoError(t, err)
		_, err = cg.cpuQuota()
		assert.NoError(t, err)
		_, err = cg.cpuUsage()
		assert.NoError(t, err)
	}

	// test cgroup v2
	if isUnifiedMode() {
		cg, err := newCGroupV2()
		assert.NoError(t, err)
		_, err = cg.effectiveCPUs()
		assert.NoError(t, err)
		_, err = cg.cpuQuota()
		assert.NoError(t, err)
		_, err = cg.cpuUsage()
		assert.NoError(t, err)
	}
}

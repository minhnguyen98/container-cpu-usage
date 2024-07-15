//go:build !linux
// +build !linux

package cgroups

func CollectCPUUsage() (float64, float64) {
	return 0, 0
}

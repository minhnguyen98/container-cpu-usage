package cgroups

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	pwd                 = mustGetWd()
	testDataPath        = filepath.Join(pwd, "testdata")
	testDataCGroupsPath = filepath.Join(testDataPath, "cgroups")
)

func mustGetWd() string {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return pwd
}

func TestReadFirstLine(t *testing.T) {
	tests := []struct {
		name            string
		paramName       string
		expectedContent string
		hasErr          bool
	}{
		{
			name:            "cpu",
			paramName:       "cpu.cfs_period_us",
			expectedContent: "100000",
			hasErr:          false,
		},
		{
			name:            "missing",
			paramName:       "cpu.stat",
			expectedContent: "",
			hasErr:          true,
		},
		{
			name:            "empty",
			paramName:       "cpu.cfs_quota_us",
			expectedContent: "",
			hasErr:          true,
		},
	}
	for _, tt := range tests {
		cgroupPath := filepath.Join(testDataCGroupsPath, tt.name, tt.paramName)

		content, err := readFirstLine(cgroupPath)
		assert.Equal(t, tt.expectedContent, content, tt.name)

		if tt.hasErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}

func TestReadInt(t *testing.T) {
	testTable := []struct {
		name          string
		paramName     string
		expectedValue int
		hasErr        bool
	}{
		{
			name:          "cpu",
			paramName:     "cpu.cfs_period_us",
			expectedValue: 100000,
			hasErr:        false,
		},
		{
			name:          "empty",
			paramName:     "cpu.cfs_quota_us",
			expectedValue: 0,
			hasErr:        true,
		},
		{
			name:          "invalid",
			paramName:     "cpu.cfs_quota_us",
			expectedValue: 0,
			hasErr:        true,
		},
		{
			name:          "absent",
			paramName:     "cpu.cfs_quota_us",
			expectedValue: 0,
			hasErr:        true,
		},
	}

	for _, tt := range testTable {
		cgroupPath := filepath.Join(testDataCGroupsPath, tt.name, tt.paramName)

		value, err := readInt(cgroupPath)
		assert.Equal(t, tt.expectedValue, value, "%s/%s", tt.name, tt.paramName)

		if tt.hasErr {
			assert.Error(t, err, tt.name)
		} else {
			assert.NoError(t, err, tt.name)
		}
	}
}

func TestParseUint(t *testing.T) {
	tests := []struct {
		input string
		want  uint64
		err   error
	}{
		{"3", 3, nil},
		{"7798", 7798, nil},
		{"-1", 0, nil},
		{"-118347827399238", 0, nil},
		{"two", 0, fmt.Errorf("parseUint: bad format: two")},
	}

	for _, tt := range tests {
		got, err := parseUint(tt.input)
		assert.Equal(t, tt.err, err)
		assert.Equal(t, tt.want, got)
	}
}

func TestParseUints(t *testing.T) {
	tests := []struct {
		input string
		want  []uint64
		err   error
	}{
		{"", nil, nil},
		{"1,2,3", []uint64{1, 2, 3}, nil},
		{"1-3", []uint64{1, 2, 3}, nil},
		{"1-3,5,7-9", []uint64{1, 2, 3, 5, 7, 8, 9}, nil},
		{"six", nil, fmt.Errorf("parseUint: bad format: six")},
		{"1-two", nil, fmt.Errorf("parseUints: bad format: 1-two")},
		{"two-3", nil, fmt.Errorf("parseUints: bad format: two-3")},
		{"three-one", nil, fmt.Errorf("parseUints: bad format: three-one")},
	}

	for _, tt := range tests {
		got, err := parseUints(tt.input)
		assert.Equal(t, tt.err, err)
		assert.Equal(t, tt.want, got)
	}
}

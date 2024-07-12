package cgroups

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// readFirstLine reads the first line from a cgroup file.
func readFirstLine(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", io.ErrUnexpectedEOF
}

// readInt parses the first line from a cgroup file as int.
func readInt(filename string) (int, error) {
	text, err := readFirstLine(filename)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(text)
}

func parseUint(s string) (uint64, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		if errors.Is(err, strconv.ErrRange) {
			return 0, nil
		}
		return 0, fmt.Errorf("cgroups: bad int format: %s", s)
	}

	if v < 0 {
		return 0, nil
	}

	return uint64(v), nil
}

// parse cpuset.cpus format: 0-7
func parseUints(val string) ([]uint64, error) {
	if val == "" {
		return nil, nil
	}

	var sets []uint64
	cols := strings.Split(val, ",")
	for _, r := range cols {
		if strings.Contains(r, "-") {
			fields := strings.SplitN(r, "-", 2)
			minimum, err := parseUint(fields[0])
			if err != nil {
				return nil, fmt.Errorf("cgroups: bad int list format: %s", val)
			}

			maximum, err := parseUint(fields[1])
			if err != nil {
				return nil, fmt.Errorf("cgroups: bad int list format: %s", val)
			}

			if maximum < minimum {
				return nil, fmt.Errorf("cgroups: bad int list format: %s", val)
			}

			for i := minimum; i <= maximum; i++ {
				sets = append(sets, i)
			}
		} else {
			v, err := parseUint(r)
			if err != nil {
				return nil, err
			}

			sets = append(sets, v)
		}
	}

	return sets, nil
}

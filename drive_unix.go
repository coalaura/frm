//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func getDiskType(path string) (string, error) {
	// Use df and grep to determine the filesystem type
	out, err := exec.Command("df", "-T", path).Output()
	if err != nil {
		return "", err
	}

	output := strings.Split(string(out), "\n")
	if len(output) < 2 {
		return "", fmt.Errorf("unexpected output from df: %s", string(out))
	}

	fields := strings.Fields(output[1])
	if len(fields) < 2 {
		return "", fmt.Errorf("unexpected output from df: %s", string(out))
	}

	return fields[1], nil
}

func calculateMaxWorkers(path string) int {
	cpus := runtime.NumCPU()

	diskType, err := getDiskType(path)
	if err != nil {
		fmt.Printf("Error determining disk type: %v\n", err)

		return cpus * 4
	}

	switch diskType {
	case "ext4", "ext3", "xfs", "btrfs": // Assume these are on SSD or fast drives
		return cpus * 4
	case "nfs", "smbfs": // Network file systems
		return cpus
	case "vfat", "ntfs": // Assume these might be slower external drives
		return cpus / 2
	default:
		return cpus * 4
	}
}

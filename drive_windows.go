//go:build windows
// +build windows

package main

import (
	"fmt"
	"path/filepath"
	"runtime"
	"syscall"
	"unsafe"
)

const (
	DRIVE_UNKNOWN     = 0
	DRIVE_NO_ROOT_DIR = 1
	DRIVE_REMOVABLE   = 2
	DRIVE_FIXED       = 3
	DRIVE_REMOTE      = 4
	DRIVE_CDROM       = 5
	DRIVE_RAMDISK     = 6
)

func getDriveType(path string) (uint32, error) {
	vol := filepath.VolumeName(path)
	if vol == "" {
		return DRIVE_UNKNOWN, fmt.Errorf("unable to determine volume for path: %s", path)
	}

	drive := vol + "\\"
	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	getDriveTypeProc := kernel32.MustFindProc("GetDriveTypeW")

	drivePtr, err := syscall.UTF16PtrFromString(drive)
	if err != nil {
		return DRIVE_UNKNOWN, err
	}

	r, _, _ := getDriveTypeProc.Call(uintptr(unsafe.Pointer(drivePtr)))

	return uint32(r), nil
}

func calculateMaxWorkers(path string) int {
	cpus := runtime.NumCPU()

	driveType, err := getDriveType(path)
	if err != nil {
		fmt.Printf("Error determining drive type: %v\n", err)

		return cpus * 2
	}

	switch driveType {
	case DRIVE_FIXED: // SSDs or HDDs
		return cpus * 2
	case DRIVE_REMOVABLE: // USB drives, etc.
		return cpus / 2
	case DRIVE_REMOTE: // Network drives
		return cpus
	case DRIVE_CDROM: // Optical drives (CD, DVD)
		return cpus / 2
	case DRIVE_RAMDISK: // Fast RAM disks
		return cpus * 4
	case DRIVE_UNKNOWN, DRIVE_NO_ROOT_DIR: // Unknown or invalid drive type
		return cpus
	default:
		return cpus * 2
	}
}

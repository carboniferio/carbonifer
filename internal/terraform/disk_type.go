package terraform

import "fmt"

// ENUM(SSD, HDD)
//
//go:generate go-enum --nocase --noprefix --marshal
type DiskType int

// UnsupportedDiskTypeError is an error that occurs when a disk type is not supported
type UnsupportedDiskTypeError struct {
	DiskType string
}

func (upe *UnsupportedDiskTypeError) Error() string {
	return fmt.Sprintf("Unsupported Disk Type: %v", upe.DiskType)
}

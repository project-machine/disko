package linux

import "machinerun.io/disko"

type RAIDControllerType string

const (
	MegaRAIDControllerType RAIDControllerType = "megaraid"
)

type RAIDController interface {
	// Type() RAIDControllerType
	GetDiskType(string) (disko.DiskType, error)
	IsSysPathRAID(string) bool
	DriverSysfsPath() string
}

package linux

import "machinerun.io/disko"

type RAIDControllerType string

const (
	MegaRAIDControllerType RAIDControllerType = "megaraid"
	SmartPqiControllerType RAIDControllerType = "smartpqi"
	MPI3MRControllerType   RAIDControllerType = "mpi3mr"
)

type RAIDController interface {
	// Type() RAIDControllerType
	GetDiskType(string) (disko.DiskType, error)
	IsSysPathRAID(string) bool
	DriverSysfsPath() string
}

// +build linux

package linux

import "github.com/anuvu/disko"

func (ls *linuxSystem) ScanPVs(filter disko.PVFilter) (disko.PVSet, error) {
	return nil, nil
}

func (ls *linuxSystem) ScanVGs(filter disko.VGFilter) (disko.VGSet, error) {
	return nil, nil
}

func (ls *linuxSystem) CreatePV(name string) (disko.PV, error) {
	return nil, nil
}

func (ls *linuxSystem) CreateVG(name string, pvs ...disko.PV) (disko.VG, error) {
	return nil, nil
}

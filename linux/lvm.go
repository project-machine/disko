package linux

import (
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/anuvu/disko"
)

// VolumeManager returns the linux implementation of disko.VolumeManager interface.
func VolumeManager() disko.VolumeManager {
	return &linuxLVM{}
}

type linuxLVM struct {
}

func (ls *linuxLVM) ScanPVs(filter disko.PVFilter) (disko.PVSet, error) {
	pvs := disko.PVSet{}

	pvdatum, err := getPvReport()
	if err != nil {
		return pvs, err
	}

	for _, pvd := range pvdatum {
		pv := pvd.toPV()
		if filter(pv) {
			pvs[pv.Name] = pv
		}
	}

	return pvs, nil
}

func (ls *linuxLVM) ScanVGs(filter disko.VGFilter) (disko.VGSet, error) {
	var vgdatum []lvmVGData
	var vgs = disko.VGSet{}
	var err error

	vgdatum, err = getVgReport()
	if err != nil {
		return vgs, err
	}

	for _, vgd := range vgdatum {
		name := vgd.Name
		vg := disko.VG{
			Name:      name,
			Size:      vgd.Size,
			FreeSpace: vgd.Free,
		}

		if !filter(vg) {
			continue
		}

		pvs, err := ls.ScanPVs(getPVFilterByName(name))
		if err != nil {
			return vgs, err
		}

		lvs, err := ls.ScanLVs(
			func(d disko.LV) bool { return d.VGName == name })
		if err != nil {
			return vgs, err
		}

		vg.PVs = pvs
		vg.Volumes = lvs

		vgs[name] = vg
	}

	return vgs, nil
}

func (ls *linuxLVM) ScanLVs(filter disko.LVFilter) (disko.LVSet, error) {
	var lvdatum []lvmLVData
	var lvs = disko.LVSet{}
	var err error

	lvdatum, err = getLvReport()
	if err != nil {
		return lvs, err
	}

	for _, lvd := range lvdatum {
		lv := lvd.toLV()

		if err != nil {
			return lvs, err
		}

		if !filter(lv) {
			continue
		}

		lvs[lv.Name] = lv
	}

	return lvs, nil
}

func (ls *linuxLVM) CreatePV(name string) (disko.PV, error) {
	nilPV := disko.PV{}

	err := runCommandSettled("lvm", "pvcreate", name)

	if err != nil {
		return nilPV, err
	}

	pvs, err := ls.ScanPVs(getPVFilterByName(name))
	if err != nil {
		return nilPV, err
	}

	if len(pvs) != 1 {
		return nilPV,
			fmt.Errorf("found %d PVs with named %s: %v", len(pvs), name, pvs)
	}

	return pvs[name], nil
}

func (ls *linuxLVM) DeletePV(pv disko.PV) error {
	return runCommandSettled("lvm", "pvremove", "--force", pv.Path)
}

func (ls *linuxLVM) HasPV(name string) bool {
	pvs, err := ls.ScanPVs(getPVFilterByName(name))
	if err != nil {
		return false
	}

	return len(pvs) != 0
}

func (ls *linuxLVM) CreateVG(name string, pvs ...disko.PV) (disko.VG, error) {
	cmd := []string{"lvm", "vgcreate", name}
	for _, p := range pvs {
		cmd = append(cmd, p.Path)
	}

	err := runCommandSettled(cmd...)
	if err != nil {
		return disko.VG{}, nil
	}

	vgSet, err := ls.ScanVGs(getVGFilterByName(name))

	if err != nil {
		return disko.VG{}, nil
	}

	return vgSet[name], nil
}

func (ls *linuxLVM) ExtendVG(vgName string, pvs ...disko.PV) error {
	cmd := []string{"lvm", "vgextend", vgName}
	for _, p := range pvs {
		cmd = append(cmd, p.Path)
	}

	err := runCommandSettled(cmd...)
	if err != nil {
		return err
	}

	return nil
}

func (ls *linuxLVM) RemoveVG(vgName string) error {
	return runCommand("lvm", "lvremove", "--force", vgName)
}

func (ls *linuxLVM) HasVG(vgName string) bool {
	vgs, err := ls.ScanVGs(getVGFilterByName(vgName))
	if err != nil {
		return false
	}

	return len(vgs) != 0
}

func (ls *linuxLVM) CryptFormat(vgName string, lvName string, key string) error {
	return runCommandStdin(
		key,
		"cryptsetup", "luksFormat", "--key-file=-", lvPath(vgName, lvName))
}

func (ls *linuxLVM) CryptOpen(vgName string, lvName string,
	decryptedName string, key string) error {
	return runCommandStdin(key,
		"cryptsetup", "open", "--type=luks", "--key-file=-",
		lvPath(vgName, lvName), decryptedName)
}

func (ls *linuxLVM) CryptClose(vgName string, lvName string,
	decryptedName string) error {
	return runCommand("cryptsetup", "close", decryptedName)
}

func (ls *linuxLVM) CreateLV(vgName string, name string, size uint64,
	lvType disko.LVType) (disko.LV, error) {
	nilLV := disko.LV{}

	if err := isRoundExtent(size); err != nil {
		return nilLV, err
	}

	if lvType == disko.THIN {
		// thin lv creation would require creating a pool
		return nilLV, fmt.Errorf("not supported. Thin LV create not implemented")
	}

	err := runCommandSettled(
		"lvm", "lvcreate", "--ignoremonitoring", "--yes", "--activate=y",
		"--setactivationskip=n", fmt.Sprintf("--size=%dB", size),
		fmt.Sprintf("--name=%s", name), vgName)

	if err != nil {
		return nilLV, err
	}

	lvs, err := ls.ScanLVs(getLVFilterByName(vgName, name))

	if err != nil {
		return nilLV, err
	}

	if len(lvs) != 1 {
		return nilLV, fmt.Errorf("found %d LVs with %s/%s", len(lvs), vgName, name)
	}

	return lvs[name], nil
}

func (ls *linuxLVM) RenameLV(vgName string, lvName string, newLvName string) error {
	return runCommandSettled("lvm", "lvrename", vgName, lvName, newLvName)
}

func (ls *linuxLVM) RemoveLV(vgName string, lvName string) error {
	return runCommandSettled(
		"lvm", "lvremove", "--force", "--force", vgLv(vgName, lvName))
}

func (ls *linuxLVM) ExtendLV(vgName string, lvName string,
	newSize uint64) error {
	if err := isRoundExtent(newSize); err != nil {
		return err
	}

	return runCommandSettled(
		"lvm", "lvextend", fmt.Sprintf("--size=%dB", newSize),
		vgLv(vgName, lvName))
}

func (ls *linuxLVM) HasLV(vgName string, name string) bool {
	lvs, err := ls.ScanLVs(getLVFilterByName(vgName, name))
	if err != nil {
		log.Panicf("Failed to scan logical volumes: %s", err)
	}

	return len(lvs) != 0
}

func getVGFilterByName(name string) disko.VGFilter {
	return func(d disko.VG) bool { return d.Name == name }
}

func getPVFilterByName(name string) disko.PVFilter {
	return func(d disko.PV) bool { return d.Name == name }
}

func getLVFilterByName(vgName string, name string) disko.LVFilter {
	return func(d disko.LV) bool { return d.Name == name && d.VGName == vgName }
}

func isRoundExtent(size uint64) error {
	if size%disko.ExtentSize == 0 {
		return nil
	}

	return fmt.Errorf("%d is not evenly divisible by extent size %d",
		size, disko.ExtentSize)
}

func (d *lvmLVData) toLV() disko.LV {
	crypt := false

	lvtype := disko.THICK

	for _, l := range strings.Split(d.raw["lv_layout"], ",") {
		if l == "thin" {
			lvtype = disko.THIN
			break
		}
	}

	if pathExists(d.Path) {
		_, _, rc := runCommandWithOutputErrorRc("cryptsetup", "isLuks", d.Path)
		if rc == 0 {
			crypt = true
		}
	}

	lv := disko.LV{
		Name:      d.Name,
		Path:      d.Path,
		VGName:    d.VGName,
		Size:      d.Size,
		Type:      lvtype,
		Encrypted: crypt,
	}

	return lv
}

func (d *lvmPVData) toPV() disko.PV {
	return disko.PV{
		Path:     d.Path,
		Name:     path.Base(d.Path),
		Size:     d.Size,
		VGName:   d.VGName,
		FreeSize: d.Free,
	}
}

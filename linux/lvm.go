// +build linux

package linux

import (
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
		pv := disko.PV{
			Path:     pvd.Path,
			Name:     path.Base(pvd.Path),
			Size:     pvd.Size,
			VGName:   pvd.VGName,
			FreeSize: pvd.Free,
		}
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
	var vgd lvmVGData
	var name string

	vgdatum, err = getVgReport()
	if err != nil {
		return vgs, err
	}

	pvHasVGName := func(p disko.PV) bool { return p.VGName == name }
	lvHasVGName := func(p disko.LV) bool { return p.VGName == name }

	for _, vgd = range vgdatum {
		name = vgd.Name
		vg := disko.VG{
			Name:      name,
			Size:      vgd.Size,
			FreeSpace: vgd.Free,
		}

		if !filter(vg) {
			continue
		}

		pvs, err := ls.ScanPVs(pvHasVGName)
		if err != nil {
			return vgs, err
		}

		lvs, err := ls.ScanLVs(lvHasVGName)
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
	var lvd lvmLVData
	var err error

	lvdatum, err = getLvReport()
	if err != nil {
		return lvs, err
	}

	for _, lvd = range lvdatum {
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
	return disko.PV{}, nil
}

func (ls *linuxLVM) DeletePV(pv disko.PV) error {
	return nil
}

func (ls *linuxLVM) HasPV(name string) bool {
	hasPVName := func(p disko.PV) bool { return p.VGName == name }

	pvs, err := ls.ScanPVs(hasPVName)
	if err != nil {
		return false
	}

	return len(pvs) != 0
}

func (ls *linuxLVM) CreateVG(name string, pvs ...disko.PV) (disko.VG, error) {
	return disko.VG{}, nil
}

func (ls *linuxLVM) ExtendVG(vgName string, pvs ...disko.PV) error {
	return nil
}

func (ls *linuxLVM) RemoveVG(vgName string) error {
	return nil
}

func (ls *linuxLVM) HasVG(vgName string) bool {
	hasVGName := func(v disko.VG) bool { return v.Name == vgName }

	vgs, err := ls.ScanVGs(hasVGName)
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
	return nil
}

func (ls *linuxLVM) CreateLV(vgName string, name string, size uint64,
	lvType disko.LVType) (disko.LV, error) {
	return disko.LV{}, nil
}

func (ls *linuxLVM) RemoveLV(vgName string, lvName string) error {
	return nil
}

func (ls *linuxLVM) ExtendLV(vgName string, lvName string,
	newSize uint64) error {
	return nil
}

func (ls *linuxLVM) HasLV(vgName string, name string) bool {
	return false
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

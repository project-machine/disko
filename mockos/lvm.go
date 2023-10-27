package mockos

import (
	"fmt"
	"path"

	"machinerun.io/disko"
)

type mockLVM struct {
	VGs     disko.VGSet `json:"vgs"`
	PVs     disko.PVSet `json:"pvs"`
	sys     disko.System
	freePVs disko.PVSet
}

// LVM return mock lvm implementation.
func LVM(sys disko.System) disko.VolumeManager {
	return &mockLVM{
		VGs:     disko.VGSet{},
		PVs:     disko.PVSet{},
		sys:     sys,
		freePVs: disko.PVSet{},
	}
}

func (lvm *mockLVM) ScanPVs(filter disko.PVFilter) (disko.PVSet, error) {
	pvs := disko.PVSet{}

	for n, pv := range lvm.PVs {
		if filter == nil || filter(pv) {
			pvs[n] = pv
		}
	}

	return pvs, nil
}

func (lvm *mockLVM) ScanVGs(filter disko.VGFilter) (disko.VGSet, error) {
	vgs := disko.VGSet{}

	for n, vg := range lvm.VGs {
		if filter == nil || filter(vg) {
			vgs[n] = vg
		}
	}

	return vgs, nil
}

func hasPartition(disks disko.DiskSet, name string) bool {
	for _, d := range disks {
		for _, p := range d.Partitions {
			if p.Name == name {
				return true
			}
		}
	}

	return false
}

func (lvm *mockLVM) CreatePV(deviceName string) (disko.PV, error) {
	disks, _ := lvm.sys.ScanAllDisks(func(d disko.Disk) bool { return true })
	d, ok := disks[deviceName]

	if !ok {
		// The device is not a disk, lets check if it is a partition.
		if !hasPartition(disks, deviceName) {
			return disko.PV{}, fmt.Errorf("disk %s does not exist", deviceName)
		}
	}

	if _, ok := lvm.PVs[deviceName]; ok {
		return disko.PV{}, fmt.Errorf("pv %s already exists", deviceName)
	}

	pv := disko.PV{
		Name:     deviceName,
		Path:     path.Join("dev", deviceName),
		Size:     d.Size,
		FreeSize: d.Size,
	}

	lvm.freePVs[pv.Name] = pv
	lvm.PVs[pv.Name] = pv

	return pv, nil
}

func (lvm *mockLVM) DeletePV(pv disko.PV) error {
	if _, ok := lvm.PVs[pv.Name]; !ok {
		return fmt.Errorf("pv %s does not exist", pv.Name)
	}

	// PV must not be used by any vg to delete
	if _, ok := lvm.freePVs[pv.Name]; !ok {
		return fmt.Errorf("pv %s is in use", pv.Name)
	}

	delete(lvm.PVs, pv.Name)
	delete(lvm.freePVs, pv.Name)

	return nil
}

func (lvm *mockLVM) HasPV(name string) bool {
	_, ok := lvm.PVs[name]
	return ok
}

func (lvm *mockLVM) CreateVG(name string, pvs ...disko.PV) (disko.VG, error) {
	if _, ok := lvm.VGs[name]; ok {
		return disko.VG{}, fmt.Errorf("vg %s already exists", name)
	}

	pvSet := disko.PVSet{}
	size := uint64(0)

	for _, pv := range pvs {
		if _, ok := lvm.freePVs[pv.Name]; !ok {
			// pv already used by some other vg
			return disko.VG{}, fmt.Errorf("pv %s already in use", pv.Name)
		}

		// delete the PV from list and add it to this vg list
		delete(lvm.freePVs, pv.Name)
		pvSet[pv.Name] = pv

		size += pv.Size
	}

	vg := disko.VG{
		Name:      name,
		Size:      size,
		Volumes:   disko.LVSet{},
		FreeSpace: size,
		PVs:       pvSet,
	}

	lvm.VGs[name] = vg

	return vg, nil
}

func (lvm *mockLVM) ExtendVG(vgName string, pvs ...disko.PV) error {
	vg, ok := lvm.VGs[vgName]
	if !ok {
		return fmt.Errorf("vg %s does not exist", vgName)
	}

	for _, pv := range pvs {
		if _, ok := lvm.freePVs[pv.Name]; !ok {
			// pv already used by some other vg
			return fmt.Errorf("pv %s already in use", pv.Name)
		}
	}

	// Delete all the added pvs from the free list
	for _, pv := range pvs {
		delete(lvm.freePVs, pv.Name)
		vg.PVs[pv.Name] = pv
		vg.Size += pv.Size
		vg.FreeSpace += pv.FreeSize
	}

	lvm.VGs[vg.Name] = vg

	return nil
}

func (lvm *mockLVM) RemoveVG(vgName string) error {
	vg, ok := lvm.VGs[vgName]
	if !ok {
		return fmt.Errorf("vg %s does not exist", vgName)
	}

	for _, pv := range vg.PVs {
		// Add all the pvs from this vg into the free list
		lvm.freePVs[pv.Name] = pv
	}

	// Delete this VG from lvm
	delete(lvm.VGs, vgName)

	return nil
}

func (lvm *mockLVM) HasVG(vgName string) bool {
	_, ok := lvm.VGs[vgName]
	return ok
}

func (lvm *mockLVM) CryptFormat(vgName string, lvName string,
	key string) error {
	vg, lv, err := lvm.findLV(vgName, lvName)
	if err != nil {
		return fmt.Errorf("lv %s does not exist", lvName)
	}

	lv.Encrypted = true
	vg.Volumes[lvName] = lv

	return nil
}

func (lvm *mockLVM) CryptOpen(vgName string, lvName string,
	decryptedName string, key string) error {
	vg, lv, err := lvm.findLV(vgName, lvName)
	if err != nil {
		return fmt.Errorf("lv %s does not exist", lvName)
	}

	if !lv.Encrypted {
		return fmt.Errorf("lv %s is not encrypted", lvName)
	}

	lv.DecryptedLVName = decryptedName
	lv.DecryptedLVPath = path.Join("/dev/mapper", decryptedName)
	vg.Volumes[lvName] = lv

	return nil
}

func (lvm *mockLVM) CryptClose(vgName string, lvName string,
	decryptedName string) error {
	vg, lv, err := lvm.findLV(vgName, lvName)
	if err != nil {
		return fmt.Errorf("lv %s does not exist", lvName)
	}

	if !lv.Encrypted {
		return fmt.Errorf("lv %s is not encrypted", lvName)
	}

	if lv.DecryptedLVName == "" || lv.DecryptedLVPath == "" {
		return fmt.Errorf("lv %s is not opened", lvName)
	}

	lv.DecryptedLVName = ""
	lv.DecryptedLVPath = ""
	vg.Volumes[lvName] = lv

	return nil
}

func (lvm *mockLVM) CreateLV(vgName string, name string, size uint64,
	lvType disko.LVType) (disko.LV, error) {
	vg, _, err := lvm.findLV(vgName, name)
	if err == nil {
		return disko.LV{}, fmt.Errorf("lv %s already exists", name)
	}

	vg, ok := lvm.VGs[vgName]
	if !ok {
		return disko.LV{}, fmt.Errorf("vg %s does not exist", vgName)
	}

	if vg.FreeSpace < size {
		return disko.LV{}, fmt.Errorf("vg %s does not have enough space", vgName)
	}

	lv := disko.LV{
		Name:      name,
		Size:      size,
		Type:      lvType,
		VGName:    vgName,
		Encrypted: false,
	}

	// Add the lv to vg and discount the freespace
	vg.Volumes[name] = lv
	vg.FreeSpace -= size

	lvm.VGs[vg.Name] = vg

	return lv, nil
}

func (lvm *mockLVM) RenameLV(vgName string, lvName string, newLvName string) error {
	vg, lv, err := lvm.findLV(vgName, lvName)
	if err != nil {
		return err
	}

	delete(vg.Volumes, lvName)

	vg.Volumes[newLvName] = lv
	lv.Name = newLvName

	return nil
}

func (lvm *mockLVM) RemoveLV(vgName string, lvName string) error {
	vg, lv, err := lvm.findLV(vgName, lvName)
	if err != nil {
		return err
	}

	// Delete the LV and reclaim the free space
	delete(vg.Volumes, lvName)
	vg.FreeSpace += lv.Size

	lvm.VGs[vg.Name] = vg

	return nil
}

func (lvm *mockLVM) ExtendLV(vgName string, lvName string,
	newSize uint64) error {
	vg, lv, err := lvm.findLV(vgName, lvName)
	if err != nil {
		return err
	}

	if newSize < lv.Size {
		return fmt.Errorf("lv size cannot be reduced")
	}

	deltaSize := newSize - lv.Size

	if vg.FreeSpace < deltaSize {
		return fmt.Errorf("vg %s does not have enough space", vg.Name)
	}

	// allocate the space from the vg to this lv
	vg.FreeSpace -= deltaSize
	lv.Size += deltaSize

	return nil
}

func (lvm *mockLVM) HasLV(vgName string, name string) bool {
	_, _, err := lvm.findLV(vgName, name)
	return err == nil
}

func (lvm *mockLVM) findLV(vgName string, lvName string) (disko.VG, disko.LV, error) {
	vg, ok := lvm.VGs[vgName]
	if !ok {
		return disko.VG{}, disko.LV{}, fmt.Errorf("vg %s not found", vgName)
	}

	lv, ok := vg.Volumes[lvName]
	if !ok {
		return disko.VG{}, disko.LV{}, fmt.Errorf("lv %s not found", lvName)
	}

	return vg, lv, nil
}

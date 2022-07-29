package linux

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/anuvu/disko"
)

const pvMetaDataSize = 128 * disko.Mebibyte
const thinPoolMetaDataSize = 1024 * disko.Mebibyte

// VolumeManager returns the linux implementation of disko.VolumeManager interface.
func VolumeManager() disko.VolumeManager {
	return &linuxLVM{}
}

type linuxLVM struct {
}

func (ls *linuxLVM) ScanPVs(filter disko.PVFilter) (disko.PVSet, error) {
	return ls.scanPVs(filter)
}

func (ls *linuxLVM) scanPVs(filter disko.PVFilter, scanArgs ...string) (disko.PVSet, error) {
	pvs := disko.PVSet{}

	pvdatum, err := getPvReport(scanArgs...)
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
	return ls.scanVGs(filter)
}

func (ls *linuxLVM) scanVGs(filter disko.VGFilter, scanArgs ...string) (disko.VGSet, error) {
	var vgdatum []lvmVGData
	var vgs = disko.VGSet{}
	var err error

	vgdatum, err = getVgReport(scanArgs...)
	if err != nil {
		return vgs, err
	}

	if len(vgdatum) == 0 {
		return vgs, err
	}

	for _, vgd := range vgdatum {
		name := vgd.Name
		vg := disko.VG{
			Name:      name,
			UUID:      vgd.UUID,
			Size:      vgd.Size,
			FreeSpace: vgd.Free,
		}

		if !filter(vg) {
			continue
		}

		vgs[name] = vg
	}

	if len(vgs) == 0 {
		return vgs, nil
	}

	fullVgs := disko.VGSet{}
	lvSetsByVG := map[string]disko.LVSet{}
	pvSetsByVG := map[string]disko.PVSet{}

	lvs, err := ls.scanLVs(func(d disko.LV) bool { return true })

	if err != nil {
		return vgs, err
	}

	for _, lv := range lvs {
		if _, ok := lvSetsByVG[lv.VGName]; ok {
			lvSetsByVG[lv.VGName][lv.Name] = lv
		} else {
			lvSetsByVG[lv.VGName] = disko.LVSet{lv.Name: lv}
		}
	}

	pvs, err := ls.scanPVs(func(d disko.PV) bool { return true })

	if err != nil {
		return vgs, err
	}

	for _, pv := range pvs {
		if _, ok := pvSetsByVG[pv.VGName]; ok {
			pvSetsByVG[pv.VGName][pv.Name] = pv
		} else {
			pvSetsByVG[pv.VGName] = disko.PVSet{pv.Name: pv}
		}
	}

	for _, vg := range vgs {
		vg.PVs = pvSetsByVG[vg.Name]
		vg.Volumes = lvSetsByVG[vg.Name]
		fullVgs[vg.Name] = vg
	}

	return fullVgs, nil
}

func (ls *linuxLVM) ScanLVs(filter disko.LVFilter) (disko.LVSet, error) {
	return ls.scanLVs(filter)
}

func (ls *linuxLVM) scanLVs(filter disko.LVFilter, scanArgs ...string) (disko.LVSet, error) {
	var lvdatum []lvmLVData
	var lvs = disko.LVSet{}
	var err error

	lvdatum, err = getLvReport(scanArgs...)
	if err != nil {
		return lvs, err
	}

	var crypt bool
	var cryptName, cryptPath string

	for _, lvd := range lvdatum {
		lv := lvd.toLV()

		if crypt, cryptName, cryptPath, err = getLuksInfo(lv.Path); err != nil {
			return lvs, err
		}

		lv.Encrypted = crypt
		if cryptName != "" {
			lv.DecryptedLVName = cryptName
			lv.DecryptedLVPath = cryptPath
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

	var err error
	var kname, path string

	if kname, path, err = getKnameAndPathForBlockDevice(name); err != nil {
		return nilPV, err
	}

	err = runCommandSettled("lvm", "pvcreate", "--zero=y",
		fmt.Sprintf("--metadatasize=%dB", pvMetaDataSize), path)

	if err != nil {
		return nilPV, err
	}

	pvs, err := ls.scanPVs(func(d disko.PV) bool { return true }, path)
	if err != nil {
		return nilPV, err
	}

	if len(pvs) != 1 {
		return nilPV,
			fmt.Errorf("found %d PVs named %s: %v", len(pvs), kname, pvs)
	}

	return pvs[kname], nil
}

func (ls *linuxLVM) DeletePV(pv disko.PV) error {
	return runCommandSettled("lvm", "pvremove", "--force", pv.Path)
}

func (ls *linuxLVM) HasPV(name string) bool {
	pvs, err := ls.scanPVs(func(d disko.PV) bool { return true }, getPathForKname(name))
	if err != nil {
		return false
	}

	return len(pvs) != 0
}

func (ls *linuxLVM) CreateVG(name string, pvs ...disko.PV) (disko.VG, error) {
	cmd := []string{"lvm", "vgcreate",
		fmt.Sprintf("--metadatasize=%dB", pvMetaDataSize),
		"--zero=y", name}

	for _, p := range pvs {
		cmd = append(cmd, p.Path)
	}

	err := runCommandSettled(cmd...)
	if err != nil {
		return disko.VG{}, err
	}

	vgSet, err := ls.scanVGs(func(d disko.VG) bool { return true }, name)

	if err != nil {
		return disko.VG{}, err
	}

	return vgSet[name], nil
}

func (ls *linuxLVM) ExtendVG(vgName string, pvs ...disko.PV) error {
	cmd := []string{"lvm", "vgextend", "--zero=y", vgName}
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
	return runCommand("lvm", "vgremove", "--force", vgName)
}

func (ls *linuxLVM) HasVG(vgName string) bool {
	vgs, err := ls.scanVGs(func(d disko.VG) bool { return true }, vgName)
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

func createLVCmd(args ...string) error {
	return runCommandSettled(
		append([]string{"lvm", "lvcreate", "--ignoremonitoring", "--yes", "--activate=y",
			"--setactivationskip=n"}, args...)...)
}

func createThinPool(name string, vgName string, size uint64, mdSize uint64) error {
	// thinpool takes up size + 2*mdSize
	// https://www.redhat.com/archives/linux-lvm/2020-October/thread.html#00016
	args := []string{}
	// if mdSize is zero, let lvcreate choose the size. That is documented as:
	//  (Pool_LV_size / Pool_LV_chunk_size * 64)
	if mdSize != 0 {
		args = append(args, fmt.Sprintf("--poolmetadatasize=%dB", mdSize))
	}

	return createLVCmd(append(args, "--zero=y", "--wipesignatures=y",
		fmt.Sprintf("--size=%dB", size), "--thinpool="+name, vgName)...)
}

func (ls *linuxLVM) CreateLV(vgName string, name string, size uint64,
	lvType disko.LVType) (disko.LV, error) {
	nilLV := disko.LV{}

	if err := isRoundExtent(size); err != nil {
		return nilLV, err
	}

	nameFlag := "--name=" + name
	sizeB := fmt.Sprintf("%dB", size)
	vglv := vgLv(vgName, name)

	// Missing cases: LVTypeUnknown
	//exhaustive:ignore
	switch lvType {
	case disko.THIN:
		// When creating THIN LV, the VG must be <vgname>/<thinLVName>
		if !strings.Contains(vgName, "/") {
			return nilLV,
				fmt.Errorf("%s: vgName input for THIN LV name in format <vgname>/thinDataName", vgName)
		}

		vglv = vgLv(strings.Split(vgName, "/")[0], name)

		// creation of thin volumes are always zero'd, and passing '--zero=y' will fail.
		if err := createLVCmd("--virtualsize="+sizeB, nameFlag, vgName); err != nil {
			return nilLV, err
		}
	case disko.THICK:
		if err := createLVCmd("--zero=y", "--wipesignatures=y", "--size="+sizeB, nameFlag, vgName); err != nil {
			return nilLV, err
		}

		if err := luks2Wipe(lvPath(vgName, name)); err != nil {
			return nilLV, err
		}
	case disko.THINPOOL:
		// When creating a THINPOOL, the name is the thin pool name.
		if err := createThinPool(name, vgName, size, thinPoolMetaDataSize); err != nil {
			return nilLV, err
		}
	}

	lvs, err := ls.scanLVs(func(d disko.LV) bool { return true }, vglv)

	if err != nil {
		return nilLV, err
	}

	if len(lvs) != 1 {
		return nilLV, fmt.Errorf("found %d LVs with %s/%s", len(lvs), vgName, name)
	}

	return lvs[name], nil
}

// luks2Wipe - wipe luks2 from a file/device.
// libblkid (used by wipefs and lvm) did not gain full wiping of luks2 metadata until 2.33.
// Wipe it more completely here.
func luks2Wipe(fpath string) error {
	const zeroLen = 64
	bufZero := make([]byte, zeroLen)

	// possible offsets for luks2 seconday headers from cryptsetup/lib/luks2/luks2.h
	offsets := []int64{
		0x04000, 0x008000, 0x010000, 0x020000, 0x40000,
		0x080000, 0x100000, 0x200000, 0x400000}

	return withLockedFile(fpath,
		func(fp *os.File, fInfo os.FileInfo) error {
			var wlen int64
			fileLen, err := fp.Seek(0, io.SeekEnd)
			if err != nil {
				return err
			}
			for _, offset := range offsets {
				wlen = zeroLen
				if offset >= fileLen {
					continue
				} else if offset > (fileLen - zeroLen) {
					wlen = fileLen - offset
				}
				if _, err := fp.Seek(offset, io.SeekStart); err != nil {
					return err
				}
				if n, err := fp.Write(bufZero[:wlen]); err != nil {
					return err
				} else if n != int(wlen) {
					return fmt.Errorf("short write on %s at offset %x. wrote %d, tried %d",
						fpath, offset, n, zeroLen)
				}
			}
			return nil
		})
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
	var err error

	if err = isRoundExtent(newSize); err != nil {
		return err
	}

	err = runCommandSettled(
		"lvm", "lvextend", fmt.Sprintf("--size=%dB", newSize),
		vgLv(vgName, lvName))

	if err != nil {
		return err
	}

	if crypt, cryptName, _, err := getLuksInfo(lvPath(vgName, lvName)); err != nil {
		return err
	} else if crypt && cryptName != "" {
		// luks device already opened, so resize it.
		if err := runCommandSettled("cryptsetup", "resize", cryptName); err != nil {
			return err
		}
	}

	return nil
}

func (ls *linuxLVM) HasLV(vgName string, name string) bool {
	lvs, err := ls.scanLVs(func(d disko.LV) bool { return true }, vgLv(vgName, name))
	if err != nil {
		log.Panicf("Failed to scan logical volumes: %s", err)
	}

	return len(lvs) != 0
}

func isRoundExtent(size uint64) error {
	if size%disko.ExtentSize == 0 {
		return nil
	}

	return fmt.Errorf("%d is not evenly divisible by extent size %d",
		size, disko.ExtentSize)
}

// chompBytes - strip one trailing newline if present.
func chompBytes(data []byte) []byte {
	l := len(data)
	if l == 0 || data[l-1] != '\n' {
		return data
	}

	return data[:l-1]
}

// getLuksInfo - get luks information for the provided block device path)
// returns:
//    crypt - boolean indicating if device is encrypted.
//    cryptName - name of crypt dev if device is open - "" if not encrypted.
//    cryptPath - path of crypt dev if device is open - "" if not encrypted.
//    error - nil unless an error occurred.
func getLuksInfo(devpath string) (bool, string, string, error) {
	crypt := false

	if !pathExists(devpath) {
		return crypt, "", "", nil
	}

	// $ cryptsetup luksUUID /dev/vg_ifc0/certs
	// a41a29c5-e375-4586-b30f-40eee4441db6
	cmd := []string{"cryptsetup", "luksUUID", devpath}
	stdout, stderr, rc := runCommandWithOutputErrorRc(cmd...)

	if rc == 1 {
		return crypt, "", "", nil
	} else if rc != 0 {
		return crypt, "", "", cmdError(cmd, stdout, stderr, rc)
	}
	// prefix looks like CRYPT-LUKS[12]-<luksUUID-without-spaces>-
	bareID := strings.ReplaceAll(string(chompBytes(stdout)), "-", "")
	luks1 := "CRYPT-LUKS1-" + bareID + "-"
	luks2 := "CRYPT-LUKS2-" + bareID + "-"

	crypt = true
	minFields := 4

	cmd = []string{"dmsetup", "table", "--concise"}
	stdout, stderr, rc = runCommandWithOutputErrorRc(cmd...)

	if rc != 0 {
		return crypt, "", "", cmdError(cmd, stdout, stderr, rc)
	}

	// dmsetup table --concise returns semi-colon delimited records that are comma separated.
	// per dmsetup(8): The representation of a device takes the form:
	//   <name>,<uuid>,<minor>,<flags>,<table>[,<table>+]
	for _, record := range strings.Split(string(chompBytes(stdout)), ";") {
		fields := strings.Split(record, ",")
		if len(fields) < minFields {
			return crypt, "", "",
				fmt.Errorf(
					"unexpected data in dmsetup table --concise. Found %d fields, expected >= %d: %s",
					len(fields), minFields, record)
		}

		if strings.HasPrefix(fields[1], luks1) || strings.HasPrefix(fields[1], luks2) {
			return crypt, fields[0], "/dev/mapper/" + fields[0], nil
		}
	}

	return crypt, "", "", nil
}

func (d *lvmLVData) toLV() disko.LV {
	lvtype := disko.THICK

	var isThin, isPool = false, false

	for _, l := range strings.Split(d.raw["lv_layout"], ",") {
		if l == "thin" {
			isThin = true
		}

		if l == "pool" {
			isPool = true
		}
	}

	if isPool {
		lvtype = disko.THINPOOL
	} else if isThin {
		lvtype = disko.THIN
	}

	lv := disko.LV{
		Name:      d.Name,
		UUID:      d.UUID,
		Path:      d.Path,
		VGName:    d.VGName,
		Size:      d.Size,
		Type:      lvtype,
		Encrypted: false,
	}

	return lv
}

func (d *lvmPVData) toPV() disko.PV {
	return disko.PV{
		Path:     d.Path,
		UUID:     d.UUID,
		Name:     path.Base(d.Path),
		Size:     d.Size,
		VGName:   d.VGName,
		FreeSize: d.Free,
	}
}

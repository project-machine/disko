package megaraid

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

const sysDriverMegaRaidSAS = "/sys/bus/pci/drivers/megaraid_sas"

// IsMegaRaidSysPath - is this sys path (udevadm info's DEVPATH) on a megaraid controller.
//  syspath will look something like
//     /devices/pci0000:3a/0000:3a:02.0/0000:3c:00.0/host0/target0:2:2/0:2:2:0/block/sdc
func IsMegaRaidSysPath(syspath string) bool {
	if !strings.HasPrefix(syspath, "/sys") {
		syspath = "/sys" + syspath
	}

	if !strings.Contains(syspath, "/host") {
		return false
	}

	fp, err := filepath.EvalSymlinks(syspath)
	if err != nil {
		fmt.Printf("seriously? %s\n", err)
		return false
	}

	for _, path := range getSysPaths() {
		if strings.HasPrefix(fp, path) {
			return true
		}
	}

	return false
}

// NameByDiskID - return the linux name (sda) for the disk with given DiskID
func NameByDiskID(id int) (string, error) {
	// given ID, we expect a single file in:
	// <sysDriverMegaRaidSAS>/0000:05:00.0/host0/target0:0:<ID>/0:0:<ID>:0/block/
	// Note: This does not work for some megaraid controllers such as SAS3508
	// See https://github.com/anuvu/disko/issues/101
	idStr := fmt.Sprintf("%d", id)
	blkDir := sysDriverMegaRaidSAS + "/*/host*/target0:0:" + idStr + "/0:0:" + idStr + ":0/block/*"
	matches, err := filepath.Glob(blkDir)

	if err != nil {
		return "", err
	}

	if len(matches) != 1 {
		return "", fmt.Errorf("found %d matches to %s", len(matches), blkDir)
	}

	return path.Base(matches[0]), nil
}

func getSysPaths() []string {
	paths := []string{}
	// sysDriverMegaRaidSAS has directory entries for each of the scsi hosts on that controller.
	//   $cd /sys/bus/pci/drivers/megaraid_sas
	//   $ for d in *; do [ -d "$d" ] || continue; echo "$d -> $( cd "$d" && pwd -P )"; done
	//    0000:3c:00.0 -> /sys/devices/pci0000:3a/0000:3a:02.0/0000:3c:00.0
	//    module -> /sys/module/megaraid_sas

	// We take a hack path and consider anything with a ":" in that dir as a host path.
	matches, err := filepath.Glob(sysDriverMegaRaidSAS + "/*:*")

	if err != nil {
		fmt.Printf("errors: %s\n", err)
		return paths
	}

	for _, p := range matches {
		fp, err := filepath.EvalSymlinks(p)

		if err == nil {
			paths = append(paths, fp)
		}
	}

	return paths
}

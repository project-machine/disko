package mockos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/anuvu/disko"
)

// System returns a mock os implementation of the disk.System interface.
func System(layout string) disko.System {
	file, err := ioutil.ReadFile(layout)
	if err != nil {
		panic(err)
	}

	sys := &mockSys{}

	if err := json.Unmarshal(file, sys); err != nil {
		panic(err)
	}

	return sys
}

type mockSys struct {
	Disks disko.DiskSet `json:"disks"`
}

func (ms *mockSys) ScanAllDisks(filter disko.DiskFilter) (disko.DiskSet, error) {
	disks := disko.DiskSet{}

	for n, d := range ms.Disks {
		if filter(d) {
			disks[n] = d
		}
	}

	return disks, nil
}

func (ms *mockSys) ScanDisks(filter disko.DiskFilter, paths ...string) (disko.DiskSet, error) {
	disks := disko.DiskSet{}

	for _, p := range paths {
		d, e := ms.ScanDisk(p)

		if e != nil {
			return nil, e
		}

		if filter(d) {
			disks[d.Name] = d
		}
	}

	return disks, nil
}

func (ms *mockSys) ScanDisk(path string) (disko.Disk, error) {
	// Find the disk from the disk set
	for _, d := range ms.Disks {
		if d.Path == path {
			return d, nil
		}
	}

	return disko.Disk{}, fmt.Errorf("disk %s not found", path)
}

func (ms *mockSys) CreatePartition(d disko.Disk, p disko.Partition) error {
	if _, ok := d.Partitions[p.Number]; ok {
		return fmt.Errorf("partition %d already exists", p.Number)
	}

	d.Partitions[p.Number] = p

	// Ignore free spaces for mock
	return nil
}

func (ms *mockSys) DeletePartition(d disko.Disk, n uint) error {
	if _, ok := d.Partitions[n]; !ok {
		return fmt.Errorf("partition %d does not exist", n)
	}

	delete(d.Partitions, n)

	// Ignore free space for mock
	return nil
}

func (ms *mockSys) Wipe(d disko.Disk) error {
	// later mate
	return nil
}

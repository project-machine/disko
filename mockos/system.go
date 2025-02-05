package mockos

import (
	"encoding/json"
	"fmt"
	"os"

	"machinerun.io/disko"
)

// System returns a mock os implementation of the disk.System interface.
func System(layout string) disko.System {
	file, err := os.ReadFile(layout)
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
		if filter == nil || filter(d) {
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
	if disk, ok := ms.Disks[d.Name]; ok {
		if _, ok := disk.Partitions[p.Number]; ok {
			return fmt.Errorf("partition %d already exists", p.Number)
		}

		disk.Partitions[p.Number] = p

		// Ignore free spaces for mock
		return nil
	}

	return fmt.Errorf("disk %s does not exist", d.Name)
}

func (ms *mockSys) CreatePartitions(d disko.Disk, pSet disko.PartitionSet) error {
	for _, p := range pSet {
		if err := ms.CreatePartition(d, p); err != nil {
			return err
		}
	}

	return nil
}

func (ms *mockSys) UpdatePartition(d disko.Disk, p disko.Partition) error {
	cur, ok := d.Partitions[p.Number]

	if !ok {
		return fmt.Errorf("partition %d did not exist on disk %s", p.Number, d)
	}

	emptyGUID := disko.GUID{}
	emptyType := disko.PartType{}
	upd := cur

	if p.Name != "" {
		upd.Name = p.Name
	}

	if p.ID != emptyGUID {
		upd.ID = p.ID
	}

	if p.Type != emptyType {
		upd.Type = p.Type
	}

	d.Partitions[p.Number] = upd

	return nil
}

func (ms *mockSys) UpdatePartitions(d disko.Disk, pSet disko.PartitionSet) error {
	for _, p := range pSet {
		if err := ms.UpdatePartition(d, p); err != nil {
			return err
		}
	}

	return nil
}

func (ms *mockSys) DeletePartition(d disko.Disk, number uint) error {
	if disk, ok := ms.Disks[d.Name]; ok {
		if _, ok := disk.Partitions[number]; !ok {
			return fmt.Errorf("partition %d does not exist", number)
		}

		delete(disk.Partitions, number)

		// Ignore free space for mock
		return nil
	}

	return fmt.Errorf("disk %s does not exist", d.Name)
}

func (ms *mockSys) Wipe(d disko.Disk) error {
	// later mate
	return nil
}

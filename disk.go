package disko

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"machinerun.io/disko/partid"
)

// DiskType enumerates supported disk types.
type DiskType int

const (
	// HDD - hard disk drive
	HDD DiskType = iota

	// SSD - solid state disk
	SSD

	// NVME - Non-volatile memory express
	NVME

	// TYPEFILE - A file on disk, not a block device.
	TYPEFILE
)

func (t DiskType) String() string {
	return []string{"HDD", "SSD", "NVME", "FILE"}[t]
}

// StringToDiskType - convert a string to a disk type.
func StringToDiskType(typeStr string) DiskType {
	kmap := map[string]DiskType{
		"HDD":  HDD,
		"SSD":  SSD,
		"NVME": NVME,
		"FILE": TYPEFILE,
	}
	if dtype, ok := kmap[typeStr]; ok {
		return dtype
	}

	return HDD
}

// MarshalJSON - Custom to marshal as a string.
func (t DiskType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON - custom to read as string or int.
func (t *DiskType) UnmarshalJSON(b []byte) error {
	var err error
	var asStr string
	var asInt int

	err = json.Unmarshal(b, &asInt)
	if err == nil {
		*t = DiskType(asInt)
		return nil
	}

	err = json.Unmarshal(b, &asStr)
	if err != nil {
		return err
	}

	dtype := StringToDiskType(asStr)
	*t = dtype

	return nil
}

// AttachmentType enumerates the type of device to which the disks are
// attached to in the system.
type AttachmentType int

const (
	// UnknownAttach - indicates an unknown attachment.
	UnknownAttach AttachmentType = iota

	// RAID - indicates that the device is attached to RAID card
	RAID

	// SCSI - indicates device is attached to scsi, but not a RAID card.
	SCSI

	// ATA - indicates that the device is attached to ATA card
	ATA

	// PCIE - indicates that the device is attached to PCIE card
	PCIE

	// USB - indicates that the device is attached to USB bus
	USB

	// VIRTIO - indicates that the device is attached to virtio.
	VIRTIO

	// IDE - indicates that the device is attached to IDE.
	IDE

	// NBD - a network block device (/dev/nbd0)
	NBD

	// LOOP - a loop device (/dev/loop0)
	LOOP

	// XENBUS - xen blkfront
	XENBUS

	// FILESYSTEM - a file on a filesystem.
	FILESYSTEM
)

func (t AttachmentType) String() string {
	return []string{"UNKNOWN", "RAID", "SCSI", "ATA", "PCIE", "USB",
		"VIRTIO", "IDE", "NBD", "LOOP", "XENBUS", "FILESYSTEM"}[t]
}

// StringToAttachmentType - Convert a string to an AttachmentType
func StringToAttachmentType(atypeStr string) AttachmentType {
	kmap := map[string]AttachmentType{
		"UNKNOWN":    UnknownAttach,
		"RAID":       RAID,
		"SCSI":       SCSI,
		"ATA":        ATA,
		"PCIE":       PCIE,
		"VIRTIO":     VIRTIO,
		"IDE":        IDE,
		"NBD":        NBD,
		"LOOP":       LOOP,
		"XENBUS":     XENBUS,
		"FILESYSTEM": FILESYSTEM,
	}

	if atype, ok := kmap[atypeStr]; ok {
		return atype
	}

	return UnknownAttach
}

// MarshalJSON - Custom to marshal as a string.
func (t AttachmentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON - reverse of the custom marshler
func (t *AttachmentType) UnmarshalJSON(b []byte) error {
	var err error
	var asStr string
	var asInt int

	err = json.Unmarshal(b, &asInt)
	if err == nil {
		*t = AttachmentType(asInt)
		return nil
	}

	err = json.Unmarshal(b, &asStr)
	if err != nil {
		return err
	}

	dtype := StringToAttachmentType(asStr)
	*t = dtype

	return nil
}

// TableType enumerates the type of device to which the disks are
// attached to in the system.
type TableType int

const (
	// TableNone - no table type
	TableNone TableType = iota

	// MBR - a Master Boot Record style partition table.
	MBR

	// GPT - a Guid Partition Table style partition table
	GPT
)

func (t TableType) String() string {
	return []string{"NONE", "MBR", "GPT"}[t]
}

// StringToTableType - Convert a string to an TableType
func StringToTableType(atypeStr string) TableType {
	kmap := map[string]TableType{
		"NONE": TableNone,
		"MBR":  MBR,
		"GPT":  GPT,
	}

	if atype, ok := kmap[atypeStr]; ok {
		return atype
	}

	return TableNone
}

// MarshalJSON - Custom to marshal as a string.
func (t TableType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON - reverse of the custom marshler
func (t *TableType) UnmarshalJSON(b []byte) error {
	var err error
	var asStr string
	var asInt int

	err = json.Unmarshal(b, &asInt)
	if err == nil {
		*t = TableType(asInt)
		return nil
	}

	err = json.Unmarshal(b, &asStr)
	if err != nil {
		return err
	}

	dtype := StringToTableType(asStr)
	*t = dtype

	return nil
}

// PartType represents a GPT Partition GUID
type PartType GUID

// DiskSet is a map of the kernel device name and the disk.
type DiskSet map[string]Disk

// Details prints the details of the disks in the disk set ina a tabular
// format.
func (ds DiskSet) Details() string {
	return ""
}

// Property - a property of a disk such
type Property string

const (
	// Ephemeral - A cloud ephemeral disk.
	Ephemeral Property = "EPHEMERAL"
)

// PropertySet - a group of properties of a disk
type PropertySet map[Property]bool

// MarshalJSON - serialize to json
func (p PropertySet) MarshalJSON() ([]byte, error) {
	keys := []string{}

	for k := range p {
		// Drop false values.
		if !p[k] {
			continue
		}

		keys = append(keys, string(k))
	}

	sort.Strings(keys)

	return json.Marshal(keys)
}

func (p PropertySet) String() string {
	keys := []string{}
	for k := range p {
		keys = append(keys, string(k))
	}

	sort.Strings(keys)

	return strings.Join(keys, ",")
}

// UnmarshalJSON - json unserialize
func (p *PropertySet) UnmarshalJSON(b []byte) error {
	s := map[string]bool{}

	err := json.Unmarshal(b, &s)
	if err == nil {
		for k, v := range s {
			// drop false values
			if !v {
				continue
			}

			(*p)[Property(k)] = v
		}

		return nil
	}

	slist := []string{}

	err = json.Unmarshal(b, &slist)
	if err != nil {
		return err
	}

	for _, k := range slist {
		(*p)[Property(k)] = true
	}

	return nil
}

// Disk wraps the disk level operations. It provides basic information
// about the disk including name, device path, size etc.
type Disk struct {
	// Name is the kernel name of the disk.
	Name string `json:"name"`

	// Path is the device path of the disk.
	Path string `json:"path"`

	// Size is the size of the disk in bytes.
	Size uint64 `json:"size"`

	// SectorSize is the sector size of the device, if its unknown or not
	// applicable it will return 0.
	SectorSize uint `json:"sectorSize"`

	// ReadOnly - cannot be written to.
	ReadOnly bool `json:"read-only"`

	// Type is the DiskType indicating the type of this disk. This value
	// can be used to determine if the disk is of a particular media type like
	// HDD, SSD or NVMe.
	Type DiskType `json:"type"`

	// Attachment is the type of storage card this disk is attached to.
	// For example: RAID, ATA or PCIE.
	Attachment AttachmentType `json:"attachment"`

	// Partitions is the set of partitions on this disk.
	Partitions PartitionSet `json:"partitions"`

	// TableType is the type of the table
	Table TableType `json:"table"`

	// Properties are a set of properties of this disk.
	Properties PropertySet `json:"properties"`

	// UdevInfo is the disk's udev information.
	UdevInfo UdevInfo `json:"udevInfo"`
}

// FreeSpacesWithMin returns a list of freespaces that are minSize long or more.
func (d *Disk) FreeSpacesWithMin(minSize uint64) []FreeSpace {
	// Stay out of the first 1Mebibyte
	// Leave 33 sectors at end (for GPT second header) and round 1MiB down.
	end := ((d.Size - uint64(d.SectorSize)*33) / Mebibyte) * Mebibyte
	used := uRanges{{0, 1*Mebibyte - 1}, {end, d.Size}}

	for _, p := range d.Partitions {
		used = append(used, uRange{p.Start, p.Last})
	}

	avail := []FreeSpace{}

	for _, g := range findRangeGaps(used, 0, d.Size) {
		if g.Size() < minSize {
			continue
		}

		avail = append(avail, FreeSpace(g))
	}

	return avail
}

// FreeSpaces returns a list of slots of free spaces on the disk. These slots can
// be used to create new partitions.
func (d *Disk) FreeSpaces() []FreeSpace {
	return d.FreeSpacesWithMin(ExtentSize)
}

func (d Disk) String() string {
	var avail uint64

	fs := d.FreeSpaces()

	for _, f := range fs {
		avail += f.Size()
	}

	mbsize := func(n uint64) string {
		if (n)%Mebibyte == 0 {
			return fmt.Sprintf("%dMiB", (n)/Mebibyte)
		}

		return fmt.Sprintf("%d", n)
	}

	return fmt.Sprintf(
		("%s (%s) Table=%s Size=%s NumParts=%d FreeSpace=%s/%d SectorSize=%d Attachment=%s Type=%s" +
			" ReadOnly=%t Props=%s"),
		d.Name, d.Path, d.Table, mbsize(d.Size), len(d.Partitions),
		mbsize(avail), len(fs), d.SectorSize,
		d.Attachment, d.Type, d.ReadOnly, d.Properties.String())
}

// Details returns the disk details as a table formatted string.
func (d Disk) Details() string {
	fss := d.FreeSpaces()
	var fsn int

	mbsize := func(n, o uint64) string {
		if (n+o)%Mebibyte == 0 {
			return fmt.Sprintf("%d MiB", (n+o)/Mebibyte)
		}

		return fmt.Sprintf("%d", n)
	}

	mbo := func(n uint64) string { return mbsize(n, 0) }
	mbe := func(n uint64) string { return mbsize(n, 1) }
	lfmt := "[%2s  %10s %10s %10s %-16s %-18s ]\n"
	buf := fmt.Sprintf(lfmt, "#", "Start", "Last", "Size", "Name", "Type")

	pNums := make([]uint, 0, len(d.Partitions))
	for n := range d.Partitions {
		pNums = append(pNums, n)
	}

	sort.Slice(pNums, func(i, j int) bool { return pNums[i] < pNums[j] })

	for _, n := range pNums {
		p := d.Partitions[n]

		if fsn < len(fss) && fss[fsn].Start < p.Start {
			buf += fmt.Sprintf(lfmt, "-", mbo(fss[fsn].Start), mbe(fss[fsn].Last), mbo(fss[fsn].Size()), "<free>", "None")
			fsn++
		}

		name := p.Name
		if name == "" {
			name = "N/A"
		}

		buf += fmt.Sprintf(lfmt,
			fmt.Sprintf("%d", p.Number), mbo(p.Start), mbe(p.Last), mbo(p.Size()), name, type2str(p.Type))
	}

	if fsn < len(fss) {
		buf += fmt.Sprintf(lfmt, "-", mbo(fss[fsn].Start), mbe(fss[fsn].Last), mbo(fss[fsn].Size()), "<free>", "N/A")
	}

	return buf
}

func type2str(pt PartType) string {
	if s, ok := partid.Text[pt]; ok {
		return s
	}

	return pt.String()
}

// UdevInfo captures the udev information about a disk.
type UdevInfo struct {
	// Name of the disk
	Name string `json:"name"`

	// SysPath is the system path of this device.
	SysPath string `json:"sysPath"`

	// Symlinks for the disk.
	Symlinks []string `json:"symLinks"`

	// Properties is udev information as a map of key, value pairs.
	Properties map[string]string `json:"properties"`
}

// PartitionSet is a map of partition number to the partition.
type PartitionSet map[uint]Partition

// Partition wraps the disk partition information.
type Partition struct {
	// Start is the offset in bytes of the start of this partition.
	Start uint64 `json:"start"`

	// Last is the last byte that is part of this partition.
	Last uint64 `json:"last"`

	// ID is the partition id.
	ID GUID `json:"id"`

	// Type is the partition type.
	Type PartType `json:"type"`

	// Name is the name of this partition.
	Name string `json:"name"`

	// Number is the number of this partition.
	Number uint `json:"number"`
}

// Size returns the size of the partition in bytes.
func (p *Partition) Size() uint64 {
	return p.Last - p.Start + 1
}

// jPartition - Partition, but for json (ids are strings)
type jPartition struct {
	Start  uint64 `json:"start"`
	Last   uint64 `json:"last"`
	ID     string `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Number uint   `json:"number"`
}

// UnmarshalJSON - unserialize from json
func (p *Partition) UnmarshalJSON(b []byte) error {
	j := jPartition{}

	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}

	id, err := StringToGUID(j.ID)
	if err != nil {
		return err
	}

	ptype, err := StringToGUID(j.Type)
	if err != nil {
		return err
	}

	p.Start = j.Start
	p.Last = j.Last
	p.ID = id
	p.Type = PartType(ptype)
	p.Name = j.Name
	p.Number = j.Number

	return nil
}

// MarshalJSON - serialize to json
func (p Partition) MarshalJSON() ([]byte, error) {
	return json.Marshal(jPartition{
		Start:  p.Start,
		Last:   p.Last,
		ID:     p.ID.String(),
		Type:   p.Type.String(),
		Name:   p.Name,
		Number: p.Number,
	})
}

func (p PartType) String() string {
	return GUIDToString(GUID(p))
}

// StringToPartType - convert a string to a partition type.
func StringToPartType(s string) (PartType, error) {
	p, err := StringToGUID(s)
	if err != nil {
		return PartType{}, err
	}

	return PartType(p), nil
}

// FreeSpace indicates a free slot on the disk with a Start and Last offset,
// where a partition can be created.
type FreeSpace struct {
	Start uint64 `json:"start"`
	Last  uint64 `json:"last"`
}

// Size returns the size of the free space, which is End - Start.
func (f *FreeSpace) Size() uint64 {
	return f.Last - f.Start + 1
}

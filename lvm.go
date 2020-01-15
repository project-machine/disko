package disko

type PV interface {
}

type VG interface {
	Name() string
	Size() uint64
	CreateLV(name string, size uint64, lvType LVType) (LV, error)
	RemoveLV(name string) error
	Volumes() LVSet
	FreeSpace() uint64
	Extend(pvs ...PV) error
	Delete() error
}

type LVSet map[string]LV

type LV interface {
	Name() string
	Size() uint64
	Type() LVType
	IsEncrypted() bool
	Remove()
	Extend(newSize uint64) error
}

type EncryptedLV interface {
	LV
	Key() string
}

type LVType int

const (
	THICK LVType = iota
	THIN
)

type VGSet map[string]VG

func (vgs VGSet) Details() string {

}

func CreateVG(devicePath ...string) (VG, error) {
	return nil, nil
}

func DeleteVG(name string) error {
	return nil
}

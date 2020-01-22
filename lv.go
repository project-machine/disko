package disko

// LVSet is a map of LV names to the LV.
type LVSet map[string]LV

// ExtentSize is extent size for lvm
const ExtentSize = 4 * Mebibyte

// LV interface wraps the lvm logical volume information and operations. A
// logical volume partitions a volume group into a slice of capacity that can
// be used a block device to create a file system.
type LV interface {
	// Name returns the name of the logical volume.
	Name() string

	// Size returns the size of the logical volume.
	Size() uint64

	// Type returns the type of logical volume.
	Type() LVType

	// CryptFormat setups up encryption for this volume using the provided key.
	CryptFormat(key string) error

	// CryptOpen opens the encrypted logical volume for use using the provided
	// key.
	CryptOpen(key string) error

	// CryptClose close the encrypted logical volume using the provided key.
	CryptClose(key string) error

	// IsEncrypted returns if the logical volume is encrypted.
	IsEncrypted() bool

	// Remove removes this LV.
	Remove() error

	// Extend expands the LV to the requested new size.
	Extend(newSize uint64) error
}

// LVType defines the type of the logical volume.
type LVType int

const (
	// THICK indicates thickly provisioned logical volume.
	THICK LVType = iota

	// THIN indicates thinly provisioned logical volume.
	THIN
)

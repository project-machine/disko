package disko

import (
	"github.com/rekby/gpt"
	uuid "github.com/satori/go.uuid"
)

// GUID - a 16 byte Globally Unique ID
type GUID [16]byte

// GenGUID - generate a random uuid and return it
func GenGUID() GUID {
	return GUID(uuid.NewV4())
}

func (g GUID) String() string {
	return GUIDToString(g)
}

// StringToGUID - convert a string to a GUID
func StringToGUID(sguid string) (GUID, error) {
	return gpt.StringToGuid(sguid)
}

// GUIDToString - turn a Guid into a string.
func GUIDToString(bguid GUID) string {
	return gpt.Guid(bguid).String()
}

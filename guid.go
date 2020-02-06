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
// https://github.com/rekby/gpt/pull/5
// nolint:gomnd
func GUIDToString(bguid GUID) string {
	byteToChars := func(b byte) (res []byte) {
		res = make([]byte, 0, 2)

		for i := 1; i >= 0; i-- {
			switch b >> uint(4*i) & 0x0F {
			case 0:
				res = append(res, '0')
			case 1:
				res = append(res, '1')
			case 2:
				res = append(res, '2')
			case 3:
				res = append(res, '3')
			case 4:
				res = append(res, '4')
			case 5:
				res = append(res, '5')
			case 6:
				res = append(res, '6')
			case 7:
				res = append(res, '7')
			case 8:
				res = append(res, '8')
			case 9:
				res = append(res, '9')
			case 10:
				res = append(res, 'A')
			case 11:
				res = append(res, 'B')
			case 12:
				res = append(res, 'C')
			case 13:
				res = append(res, 'D')
			case 14:
				res = append(res, 'E')
			case 15:
				res = append(res, 'F')
			}
		}

		return
	}
	s := make([]byte, 0, 36)
	byteOrder := [...]int{3, 2, 1, 0, -1, 5, 4, -1, 7, 6, -1, 8, 9, -1, 10, 11, 12, 13, 14, 15}

	for _, i := range byteOrder {
		if i == -1 {
			s = append(s, '-')
		} else {
			s = append(s, byteToChars(bguid[i])...)
		}
	}

	return string(s)
}

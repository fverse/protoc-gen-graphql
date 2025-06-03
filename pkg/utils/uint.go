package utils

import "strconv"

// Parses a string as uint32
func ParseUint32(s string) uint32 {
	val, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(val)
}

// Cast a uint32 as string
func CastUit32ToString(s uint32) string {
	return strconv.FormatUint(uint64(s), 10)
}

// Compare string and int
func CompareStringInt(s string, i int) bool {
	return s == strconv.Itoa(i)
}

package rdp

import (
	"fmt"
	"strconv"
	"strings"
)

type Version uint32

const (
	ZeroVersion = Version(0)
)

func NewVersion(major, minor byte, build uint16) (res Version) {
	return Version(uint32(major)<<24 | uint32(minor)<<16 | uint32(build))
}

func FromString(str string) Version {
	parts := strings.Split(str, ".")
	var version uint64

	for i, v := range parts {
		if val, err := strconv.Atoi(v); err == nil {
			switch i {
			case 1:
				version = version << 8
			case 2:
				version = version << 16
			}
			version = version | uint64(val)
		}
	}
	return Version(version)
}

func (v Version) Major() byte {
	return byte(0xff & (v >> 24))
}

func (v Version) Minor() byte {
	return byte(0xff & (v >> 16))
}

func (v Version) Build() uint16 {
	return uint16(0xffff & v)
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Build())
}

func (v Version) CompareByMinor(v2 Version) int {
	minor1 := v >> 16
	minor2 := v2 >> 16
	if minor1 > minor2 {
		return 1
	} else if minor1 < minor2 {
		return -1
	}
	return 0
}

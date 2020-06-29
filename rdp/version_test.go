package rdp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVersion(t *testing.T) {
	v := NewVersion(1, 2, 12000)
	assert.Equal(t, byte(1), v.Major())
	assert.Equal(t, byte(2), v.Minor())
	assert.Equal(t, uint16(12000), v.Build())

	v2 := FromString("4.10.159")
	assert.Equal(t, byte(4), v2.Major())
	assert.Equal(t, byte(10), v2.Minor())
	assert.Equal(t, uint16(159), v2.Build())
}

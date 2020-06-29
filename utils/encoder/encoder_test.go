package encoder

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testStruct struct {
	Val1 int
	Val2 string
}

func TestEncoder(t *testing.T) {
	value := testStruct{
		Val1: 1,
		Val2: "1",
	}
	str := EncodeRaw(value)
	fmt.Println(str)
	val2 := testStruct{}
	err := DecodeRaw(str, &val2)
	assert.NoError(t, err)
	assert.Equal(t, value.Val1, val2.Val1)
	assert.Equal(t, value.Val2, val2.Val2)
}

package hl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytesToHL(t *testing.T) {
	input := []byte{0x01, 0x02, 0x03}
	expected := "0x01:0x02:0x03"
	actual := bytesToHL(input)
	assert.Equal(t, expected, actual)
}

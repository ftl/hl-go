package hl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRigInfoVFO(t *testing.T) {
	input := "VFOA Freq=145000000 Mode=FM Width=15000 RX=1 TX=1"
	expected := RigInfoVFO{
		VFO:       VFOA,
		Frequency: 145_000_000,
		Mode:      ModeFM,
		Passband:  15_000,
		RXActive:  true,
		TXActive:  true,
	}
	actual, err := parseRigInfoVFO(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestParseRigInfo(t *testing.T) {
	input := "get_rig_info: currVFO\nVFO=VFOA Freq=145000000 Mode=FM Width=15000 RX=1 TX=1\nVFO=VFOB Freq=146000000 Mode=FM Width=15000 RX=0 TX=0\nSplit=0 SatMode=0\nRig=Dummy\nApp=BLAH\nVersion=20241103 1.1.0\nModel=1\nCRC=0x00690aee\n\nRPRT 0\n"
	expected := RigInfo{
		VFOs: []RigInfoVFO{
			{VFOA, 145_000_000, ModeFM, 15_000, true, true},
			{VFOB, 146_000_000, ModeFM, 15_000, false, false},
		},
		SplitActive:   false,
		SATModeActive: false,
		Rig:           "Dummy",
		App:           "BLAH",
		Version:       "20241103 1.1.0",
		Model:         "1",
		CRC:           "0x00690aee",
	}
	actual, err := parseRigInfo(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

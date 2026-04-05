package hl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseModeModes(t *testing.T) {
	input := "Modes: AM CW USB LSB RTTY FM WFM CWR CW-R RTTYR RTTY-R\n" +
		"Bandwidths:\n" +
		"\tAM\tNormal: 8.0000 kHz,\tNarrow: 2.4000 kHz,\tWide: 10.0000 kHz\n" +
		"\tCW\tNormal: 500.0 Hz,\tNarrow: 50.0 Hz,\tWide: 2.4000 kHz\n" +
		"\tUSB\tNormal: 2.4000 kHz,\tNarrow: 1.8000 kHz,\tWide: 3.0000 kHz\n" +
		"\tLSB\tNormal: 2.4000 kHz,\tNarrow: 1.8000 kHz,\tWide: 3.0000 kHz\n" +
		"\tRTTY\tNormal: 300.0 Hz,\tNarrow: 50.0 Hz,\tWide: 2.4000 kHz\n" +
		"\tFM\tNormal: 15.0000 kHz,\tNarrow: 8.0000 kHz,\tWide: 0.0 Hz\n" +
		"\tWFM\tNormal: 230.0000 kHz,\tNarrow: 0.0 Hz,\tWide: 0.0 Hz\n" +
		"\tCWR\tNormal: 500.0 Hz,\tNarrow: 0.0 Hz,\tWide: 0.0 Hz\n" +
		"\tRTTYR\tNormal: 300.0 Hz,\tNarrow: 0.0 Hz,\tWide: 0.0 Hz"

	result, err := parseModes(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		mode   Mode
		normal Frequency
		narrow Frequency
		wide   Frequency
	}{
		{ModeAM, 8000, 2400, 10000},
		{ModeCW, 500, 50, 2400},
		{ModeUSB, 2400, 1800, 3000},
		{ModeRTTY, 300, 50, 2400},
		{ModeFM, 15000, 8000, 0},
		{ModeWFM, 230000, 0, 0},
		{ModeCWR, 500, 0, 0},
		{ModeRTTYR, 300, 0, 0},
	}

	for _, test := range tests {
		t.Run(string(test.mode), func(t *testing.T) {
			actual, ok := result[test.mode]
			assert.True(t, ok)
			assert.Equal(t, test.normal, actual.Normal, "normal")
			assert.Equal(t, test.narrow, actual.Narrow, "narrow")
			assert.Equal(t, test.wide, actual.Wide, "wide")
		})
	}
}

func TestParseBandwidth(t *testing.T) {
	tests := []struct {
		input    string
		expected Frequency
		invalid  bool
	}{
		{
			input:    "300.0 Hz",
			expected: 300,
		},
		{
			input:    "300Hz",
			expected: 300,
		},
		{
			input:    "2.400 kHz",
			expected: 2_400,
		},
		{
			input:    "2400Hz",
			expected: 2_400,
		},
		{
			input:    "1.0 MHz",
			expected: 1_000_000,
		},
		{
			input:   "invalid",
			invalid: true,
		},
		{
			input:   "300.0 kc",
			invalid: true,
		},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			actual, err := parseBandwidth(test.input)
			if test.invalid {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

package hl

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseResponses(t *testing.T) {
	tests := []struct {
		name     string
		parse    ResponseParser
		input    string
		expected Response
		invalid  bool
	}{
		{
			name:  "get_freq",
			parse: parseRegularResponse,
			input: "get_freq:\nFrequency: 10101000\nRPRT 0\n",
			expected: Response{
				CommandEcho: "get_freq:",
				Data:        map[string]string{"Frequency": "10101000"},
				ReturnCode:  RigOk,
			},
		},
		{
			name:  "get_freq with VFO",
			parse: parseRegularResponse,
			input: "get_freq: VFOA\nFrequency: 10101000\nRPRT 0\n",
			expected: Response{
				CommandEcho: "get_freq: VFOA",
				Data:        map[string]string{"Frequency": "10101000"},
				ReturnCode:  RigOk,
			},
		},
		{
			name:  "set_freq",
			parse: parseRegularResponse,
			input: "set_freq: 10101000\nRPRT 0\n",
			expected: Response{
				CommandEcho: "set_freq: 10101000",
				Data:        map[string]string{},
				ReturnCode:  RigOk,
			},
		},
		{
			name:  "set_freq with VFOA",
			parse: parseRegularResponse,
			input: "set_freq: VFOA:10101000\nRPRT 0\n",
			expected: Response{
				CommandEcho: "set_freq: VFOA:10101000",
				Data:        map[string]string{},
				ReturnCode:  RigOk,
			},
		},
		{
			name:  "single value single line",
			parse: parseSingleValue,
			input: "get_func: VFOA:SPECTRUM\n0\nRPRT 0\n",
			expected: Response{
				CommandEcho: "get_func: VFOA:SPECTRUM",
				Data:        map[string]string{"Value": "0"},
				ReturnCode:  RigOk,
			},
		},
		{
			name:  "single value multiple lines",
			parse: parseSingleValue,
			input: "hamlib_version: currVFO\nrigctl(d), Hamlib 4.6.2 2025-02-09T21:03:50Z SHA=870364 64-bit\n\nCopyright (C) 2000-2012 Stephane Fillod\nCopyright (C) 2000-2003 Frank Singleton\nThis is free software; see the source for copying conditions.  There is NO\nwarranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.\nRPRT 0\n",
			expected: Response{
				CommandEcho: "hamlib_version: currVFO",
				Data:        map[string]string{"Value": "rigctl(d), Hamlib 4.6.2 2025-02-09T21:03:50Z SHA=870364 64-bit\n\nCopyright (C) 2000-2012 Stephane Fillod\nCopyright (C) 2000-2003 Frank Singleton\nThis is free software; see the source for copying conditions.  There is NO\nwarranty; not even for MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE."},
				ReturnCode:  RigOk,
			},
		},
		{
			name:  "single value multiple lines with crippled report line",
			parse: parseSingleValue,
			input: "get_mode_bandwidths: currVFO:CW\nMode=CW\nNormal=500Hz\nNarrow=50Hz\nWide=2400HzRPRT 0\n",
			expected: Response{
				CommandEcho: "get_mode_bandwidths: currVFO:CW",
				Data:        map[string]string{"Value": "Mode=CW\nNormal=500Hz\nNarrow=50Hz\nWide=2400Hz"},
				ReturnCode:  RigOk,
			},
		},
		{
			name:  "invalid regular response",
			parse: parseRegularResponse,
			input: "invalid:\nRPRT -1\n",
			expected: Response{
				CommandEcho: "invalid:",
				Data:        map[string]string{},
				ReturnCode:  RigInvalidParameter,
			},
		},
		{
			name:  "invalid single value",
			parse: parseSingleValue,
			input: "invalid:\nRPRT -1\n",
			expected: Response{
				CommandEcho: "invalid:",
				Data:        map[string]string{},
				ReturnCode:  RigInvalidParameter,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buffer := bytes.NewBufferString(test.input)
			reader := bufio.NewReader(buffer)
			actual, err := test.parse(reader)
			if test.invalid {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

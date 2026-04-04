package hl

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Response
		invalid  bool
	}{
		{
			name:  "get_freq",
			input: "get_freq:\nFrequency: 10101000\nRPRT 0\n",
			expected: Response{
				CommandEcho: "get_freq:",
				Data:        map[string]string{"Frequency": "10101000"},
				ReturnCode:  RigOk,
			},
		},
		{
			name:  "get_freq with VFO",
			input: "get_freq: VFOA\nFrequency: 10101000\nRPRT 0\n",
			expected: Response{
				CommandEcho: "get_freq: VFOA",
				Data:        map[string]string{"Frequency": "10101000"},
				ReturnCode:  RigOk,
			},
		},
		{
			name:  "set_freq",
			input: "set_freq: 10101000\nRPRT 0\n",
			expected: Response{
				CommandEcho: "set_freq: 10101000",
				Data:        map[string]string{},
				ReturnCode:  RigOk,
			},
		},
		{
			name:  "set_freq with VFOA",
			input: "set_freq: VFOA:10101000\nRPRT 0\n",
			expected: Response{
				CommandEcho: "set_freq: VFOA:10101000",
				Data:        map[string]string{},
				ReturnCode:  RigOk,
			},
		},
		{
			name:  "invalid",
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
			actual, err := parseResponse(reader)
			if test.invalid {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

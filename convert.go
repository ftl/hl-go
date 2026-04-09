package hl

import (
	"fmt"
	"strings"
)

func boolToHL(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func frequencyToHL(f Frequency) string {
	return fmt.Sprintf("%0.0f", f)
}

func bytesToHL(bytes []byte) string {
	parts := make([]string, len(bytes))
	for i := range bytes {
		parts[i] = fmt.Sprintf("0x%02x", bytes[i])
	}
	return strings.Join(parts, ":")
}

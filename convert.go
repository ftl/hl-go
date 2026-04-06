package hl

import "fmt"

func boolToHL(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func frequencyToHL(f Frequency) string {
	return fmt.Sprintf("%0.0f", f)
}

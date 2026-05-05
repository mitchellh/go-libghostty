// Example: modes demonstrates the Mode API from libghostty.
// It prints the value, ANSI flag, and packed hex for a couple of modes.
package main

import (
	"fmt"

	ghostty "go.mitchellh.com/libghostty"
)

func main() {
	// DEC mode 25: cursor visible (DECTCEM)
	m := ghostty.ModeCursorVisible
	fmt.Printf("value=%d ansi=%v packed=0x%04x\n", m.Value(), m.ANSI(), uint16(m))

	// ANSI mode 4: insert mode
	m = ghostty.ModeInsert
	fmt.Printf("value=%d ansi=%v packed=0x%04x\n", m.Value(), m.ANSI(), uint16(m))
}

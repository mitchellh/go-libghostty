package libghostty

// Mouse event representation and manipulation.
// Wraps the C APIs from mouse/event.h.

/*
#include <ghostty/vt.h>
*/
import "C"

import (
	"encoding"
	"fmt"
)

// Compile-time assertions that MouseButton implements the standard
// text marshaling interfaces.
var (
	_ encoding.TextMarshaler   = MouseButton(0)
	_ encoding.TextUnmarshaler = (*MouseButton)(nil)
)

// MouseEvent is an opaque handle representing a normalized mouse
// input event containing action, button, modifiers, and surface-space
// position. It is mutable and reusable, but not safe for concurrent
// use.
//
// C: GhosttyMouseEvent
type MouseEvent struct {
	ptr C.GhosttyMouseEvent
}

// MouseAction represents the type of mouse event (press, release,
// or motion).
//
// C: GhosttyMouseAction
type MouseAction int

const (
	// MouseActionPress indicates a mouse button was pressed.
	MouseActionPress MouseAction = C.GHOSTTY_MOUSE_ACTION_PRESS

	// MouseActionRelease indicates a mouse button was released.
	MouseActionRelease MouseAction = C.GHOSTTY_MOUSE_ACTION_RELEASE

	// MouseActionMotion indicates the mouse moved.
	MouseActionMotion MouseAction = C.GHOSTTY_MOUSE_ACTION_MOTION
)

// MouseButton identifies which mouse button was involved in an event.
//
// C: GhosttyMouseButton
type MouseButton int

const (
	MouseButtonUnknown MouseButton = C.GHOSTTY_MOUSE_BUTTON_UNKNOWN
	MouseButtonLeft    MouseButton = C.GHOSTTY_MOUSE_BUTTON_LEFT
	MouseButtonRight   MouseButton = C.GHOSTTY_MOUSE_BUTTON_RIGHT
	MouseButtonMiddle  MouseButton = C.GHOSTTY_MOUSE_BUTTON_MIDDLE
	MouseButtonFour    MouseButton = C.GHOSTTY_MOUSE_BUTTON_FOUR
	MouseButtonFive    MouseButton = C.GHOSTTY_MOUSE_BUTTON_FIVE
	MouseButtonSix     MouseButton = C.GHOSTTY_MOUSE_BUTTON_SIX
	MouseButtonSeven   MouseButton = C.GHOSTTY_MOUSE_BUTTON_SEVEN
	MouseButtonEight   MouseButton = C.GHOSTTY_MOUSE_BUTTON_EIGHT
	MouseButtonNine    MouseButton = C.GHOSTTY_MOUSE_BUTTON_NINE
	MouseButtonTen     MouseButton = C.GHOSTTY_MOUSE_BUTTON_TEN
	MouseButtonEleven  MouseButton = C.GHOSTTY_MOUSE_BUTTON_ELEVEN
)

// mouseButtonNames is the canonical mapping between MouseButton
// values and their snake_case string names. Used as the source of
// truth for both String and FromString.
var mouseButtonNames = []struct {
	button MouseButton
	name   string
}{
	{MouseButtonUnknown, "unknown"},
	{MouseButtonLeft, "left"},
	{MouseButtonRight, "right"},
	{MouseButtonMiddle, "middle"},
	{MouseButtonFour, "four"},
	{MouseButtonFive, "five"},
	{MouseButtonSix, "six"},
	{MouseButtonSeven, "seven"},
	{MouseButtonEight, "eight"},
	{MouseButtonNine, "nine"},
	{MouseButtonTen, "ten"},
	{MouseButtonEleven, "eleven"},
}

// String returns the canonical lowercase name of the mouse button
// (e.g. "left", "right", "four"). Unknown values render as
// "unknown".
func (b MouseButton) String() string {
	for _, e := range mouseButtonNames {
		if e.button == b {
			return e.name
		}
	}
	return "unknown"
}

// MarshalText implements encoding.TextMarshaler. The output is the
// same as String() so the type integrates with encoding/json and
// other text-based encoders.
func (b MouseButton) MarshalText() ([]byte, error) {
	return []byte(b.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler. It parses a
// canonical mouse button name (e.g. "left", "right", "four") and
// stores the corresponding MouseButton value in the receiver.
// Returns an error if the name is not recognized.
func (b *MouseButton) UnmarshalText(text []byte) error {
	s := string(text)
	for _, e := range mouseButtonNames {
		if e.name == s {
			*b = e.button
			return nil
		}
	}
	return fmt.Errorf("libghostty: unknown mouse button %q", s)
}

// ParseMouseButton returns the MouseButton value for the given
// canonical name (e.g. "left", "right", "four"). Returns an error
// if the name is not recognized.
func ParseMouseButton(s string) (MouseButton, error) {
	var b MouseButton
	if err := b.UnmarshalText([]byte(s)); err != nil {
		return MouseButtonUnknown, err
	}
	return b, nil
}

// MousePosition represents a mouse position in surface-space pixels.
//
// C: GhosttyMousePosition
type MousePosition struct {
	X float32
	Y float32
}

// NewMouseEvent creates a new mouse event with default values. The
// event must be freed with Close when no longer needed.
func NewMouseEvent() (*MouseEvent, error) {
	var ptr C.GhosttyMouseEvent
	if err := resultError(C.ghostty_mouse_event_new(nil, &ptr)); err != nil {
		return nil, err
	}
	return &MouseEvent{ptr: ptr}, nil
}

// Close frees the underlying mouse event handle. After this call,
// the mouse event must not be used.
func (e *MouseEvent) Close() {
	C.ghostty_mouse_event_free(e.ptr)
}

// SetAction sets the mouse action (press, release, motion).
func (e *MouseEvent) SetAction(action MouseAction) {
	C.ghostty_mouse_event_set_action(e.ptr, C.GhosttyMouseAction(action))
}

// Action returns the mouse action (press, release, motion).
func (e *MouseEvent) Action() MouseAction {
	return MouseAction(C.ghostty_mouse_event_get_action(e.ptr))
}

// SetButton sets the mouse button for the event.
func (e *MouseEvent) SetButton(button MouseButton) {
	C.ghostty_mouse_event_set_button(e.ptr, C.GhosttyMouseButton(button))
}

// ClearButton clears the event button, setting it to "none".
// Use this for motion events with no button pressed.
func (e *MouseEvent) ClearButton() {
	C.ghostty_mouse_event_clear_button(e.ptr)
}

// Button returns the mouse button and whether one is set. If no
// button is set (e.g. a motion-only event), ok is false.
func (e *MouseEvent) Button() (button MouseButton, ok bool) {
	var b C.GhosttyMouseButton
	ok = bool(C.ghostty_mouse_event_get_button(e.ptr, &b))
	return MouseButton(b), ok
}

// SetMods sets the keyboard modifiers held during the mouse event.
func (e *MouseEvent) SetMods(mods Mods) {
	C.ghostty_mouse_event_set_mods(e.ptr, C.GhosttyMods(mods))
}

// Mods returns the keyboard modifiers held during the mouse event.
func (e *MouseEvent) Mods() Mods {
	return Mods(C.ghostty_mouse_event_get_mods(e.ptr))
}

// SetPosition sets the event position in surface-space pixels.
func (e *MouseEvent) SetPosition(pos MousePosition) {
	C.ghostty_mouse_event_set_position(e.ptr, C.GhosttyMousePosition{
		x: C.float(pos.X),
		y: C.float(pos.Y),
	})
}

// Position returns the event position in surface-space pixels.
func (e *MouseEvent) Position() MousePosition {
	p := C.ghostty_mouse_event_get_position(e.ptr)
	return MousePosition{X: float32(p.x), Y: float32(p.y)}
}

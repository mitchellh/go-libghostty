package libghostty

// Render state for creating high-performance renderers.
// Wraps the GhosttyRenderState C APIs (excluding row/cell iterators).

/*
#include <ghostty/vt.h>
*/
import "C"

// RenderState holds the state required to render a visible screen
// (viewport) of a terminal instance. It is stateful and optimized
// for repeated updates from a single terminal, only updating dirty
// regions of the screen.
//
// A render state owns its own snapshot storage. Hold exclusive access
// to the terminal while calling [RenderState.Update]. After Update
// returns, the render state can be read without touching the terminal
// until the next Update. Do not call Update concurrently with reads
// from the same render state.
//
// Iterators populated from the render state are only valid until the
// next Update, but copied values returned from their getter methods can
// be retained.
//
// Basic usage:
//  1. Create an empty render state with NewRenderState.
//  2. Update it from a terminal via Update whenever needed.
//  3. Read from the render state to get data for drawing.
//
// C: GhosttyRenderState
type RenderState struct {
	ptr C.GhosttyRenderState
}

// RenderStateDirty describes the dirty state after an update.
// C: GhosttyRenderStateDirty
type RenderStateDirty int

const (
	// RenderStateDirtyFalse means not dirty; rendering can be skipped.
	RenderStateDirtyFalse RenderStateDirty = C.GHOSTTY_RENDER_STATE_DIRTY_FALSE

	// RenderStateDirtyPartial means some rows changed; renderer can
	// redraw incrementally.
	RenderStateDirtyPartial RenderStateDirty = C.GHOSTTY_RENDER_STATE_DIRTY_PARTIAL

	// RenderStateDirtyFull means global state changed; renderer should
	// redraw everything.
	RenderStateDirtyFull RenderStateDirty = C.GHOSTTY_RENDER_STATE_DIRTY_FULL
)

// CursorVisualStyle describes the visual style of the cursor.
// C: GhosttyRenderStateCursorVisualStyle
type CursorVisualStyle int

const (
	// CursorVisualStyleBar is a bar cursor (DECSCUSR 5, 6).
	CursorVisualStyleBar CursorVisualStyle = C.GHOSTTY_RENDER_STATE_CURSOR_VISUAL_STYLE_BAR

	// CursorVisualStyleBlock is a block cursor (DECSCUSR 1, 2).
	CursorVisualStyleBlock CursorVisualStyle = C.GHOSTTY_RENDER_STATE_CURSOR_VISUAL_STYLE_BLOCK

	// CursorVisualStyleUnderline is an underline cursor (DECSCUSR 3, 4).
	CursorVisualStyleUnderline CursorVisualStyle = C.GHOSTTY_RENDER_STATE_CURSOR_VISUAL_STYLE_UNDERLINE

	// CursorVisualStyleBlockHollow is a hollow block cursor.
	CursorVisualStyleBlockHollow CursorVisualStyle = C.GHOSTTY_RENDER_STATE_CURSOR_VISUAL_STYLE_BLOCK_HOLLOW
)

// RenderStateColors holds all color information from a render state,
// retrieved in a single call via the sized-struct API.
// C: GhosttyRenderStateColors
type RenderStateColors struct {
	// Background is the default/current background color.
	Background ColorRGB

	// Foreground is the default/current foreground color.
	Foreground ColorRGB

	// Cursor is the cursor color when explicitly set by terminal state.
	// Only valid when CursorHasValue is true.
	Cursor ColorRGB

	// CursorHasValue is true when Cursor contains a valid explicit
	// cursor color value.
	CursorHasValue bool

	// Palette is the active 256-color palette.
	Palette Palette
}

// NewRenderState creates a new empty render state.
func NewRenderState() (*RenderState, error) {
	var ptr C.GhosttyRenderState
	if err := resultError(C.ghostty_render_state_new(nil, &ptr)); err != nil {
		return nil, err
	}
	return &RenderState{ptr: ptr}, nil
}

// Close frees the underlying render state handle. After this call,
// the render state must not be used.
func (rs *RenderState) Close() {
	C.ghostty_render_state_free(rs.ptr)
}

// Update updates the render state from a terminal instance. This
// consumes terminal/screen dirty state and is the only render-state
// operation that touches the terminal. Hold exclusive access to the
// terminal while this call is running, and do not read from the same
// render state concurrently with Update.
func (rs *RenderState) Update(t *Terminal) error {
	return resultError(C.ghostty_render_state_update(rs.ptr, t.ptr))
}

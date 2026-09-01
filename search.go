package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

// Search finds a byte sequence in a terminal's active screen and scrollback
// history. It preserves results for both the primary and alternate screens.
// Calls to [Search.Feed] update those results as the terminal changes. Search
// work can be split across calls so that large scrollback histories do not
// block the caller.
//
// A Search uses, but does not own, the terminal passed to [NewSearch]. The
// search and terminal may be closed in either order.
//
// A Search is not safe for concurrent use. [Search.Tick], [Search.GetMulti],
// the data-reading methods, and [Search.SetScrollPolicy] do not access the
// terminal. These methods may run while another goroutine uses the terminal,
// provided no other goroutine is using the same Search. All other Search
// operations must be serialized with every other operation on the terminal.
// Before passing a returned [Selection] to a [Terminal] method, ensure the
// terminal has not changed since the selection was read and hold exclusive
// access to the terminal for that call.
//
// C: GhosttySearch
type Search struct {
	ptr C.GhosttySearch
}

// SearchStatus describes whether a search can make progress without reading
// from its terminal.
//
// C: GhosttySearchStatus
type SearchStatus int

const (
	// SearchStatusRunning means [Search.Tick] can make progress using data that
	// the search has already copied from the terminal.
	SearchStatusRunning SearchStatus = C.GHOSTTY_SEARCH_STATUS_RUNNING

	// SearchStatusFeedRequired means [Search.Feed] must copy more terminal data
	// before [Search.Tick] can continue.
	SearchStatusFeedRequired SearchStatus = C.GHOSTTY_SEARCH_STATUS_FEED_REQUIRED

	// SearchStatusComplete means the search is caught up as of the most recent
	// [Search.Feed]. An idle Search is also complete. Call [Search.Feed] again
	// after the terminal changes.
	SearchStatusComplete SearchStatus = C.GHOSTTY_SEARCH_STATUS_COMPLETE
)

// SearchScrollPolicy controls whether selecting a match moves the terminal
// viewport.
//
// C: GhosttySearchScroll
type SearchScrollPolicy int

const (
	// SearchScrollPolicyIfNeeded moves the viewport only when the selected
	// match is not visible. This is the default policy.
	SearchScrollPolicyIfNeeded SearchScrollPolicy = C.GHOSTTY_SEARCH_SCROLL_IF_NEEDED

	// SearchScrollPolicyNever leaves the viewport unchanged when a match is
	// selected.
	SearchScrollPolicyNever SearchScrollPolicy = C.GHOSTTY_SEARCH_SCROLL_NONE
)

// SearchData identifies a low-level field that can be read with
// [Search.GetMulti]. Most callers should use methods such as [Search.Status]
// and [Search.Matches] instead.
//
// C: GhosttySearchData
type SearchData int

const (
	// SearchDataStatus is the current search status (an int32 value compatible
	// with GhosttySearchStatus).
	SearchDataStatus SearchData = C.GHOSTTY_SEARCH_DATA_STATUS

	// SearchDataNeedle is the current search needle (GhosttyString). Reading this
	// field returns an error with result [ResultNoValue] when the Search is idle.
	// Prefer [Search.Needle].
	SearchDataNeedle SearchData = C.GHOSTTY_SEARCH_DATA_NEEDLE

	// SearchDataMatchCount is the number of matches found so far on the active
	// screen, including scrollback (uint).
	SearchDataMatchCount SearchData = C.GHOSTTY_SEARCH_DATA_TOTAL_MATCHES

	// SearchDataSelectedIndex is the selected match index (uint). Reading this
	// field returns an error with result [ResultNoValue] when no match is
	// selected.
	SearchDataSelectedIndex SearchData = C.GHOSTTY_SEARCH_DATA_SELECTED_INDEX

	// SearchDataSelectedMatch is the selected match (GhosttySelection). Reading
	// this field returns an error with result [ResultNoValue] when no match is
	// selected. Prefer [Search.SelectedMatch].
	SearchDataSelectedMatch SearchData = C.GHOSTTY_SEARCH_DATA_SELECTED_MATCH

	// SearchDataMatches contains the matches found so far on the active screen,
	// ordered from newest to oldest (GhosttySelectionBuffer). Prefer
	// [Search.Matches].
	SearchDataMatches SearchData = C.GHOSTTY_SEARCH_DATA_MATCHES

	// SearchDataViewportMatches contains matches near the viewport
	// (GhosttySelectionBuffer). Prefer [Search.ViewportMatches].
	SearchDataViewportMatches SearchData = C.GHOSTTY_SEARCH_DATA_VIEWPORT_MATCHES

	// SearchDataScrollPolicy is the current scroll policy (an int32 value
	// compatible with GhosttySearchScroll).
	SearchDataScrollPolicy SearchData = C.GHOSTTY_SEARCH_DATA_SELECT_SCROLL
)

// NewSearch creates an idle search for the given terminal. Call
// [Search.SetNeedle] to start searching, and call [Search.Close] when the
// search is no longer needed.
//
// NewSearch registers with the terminal so that either value can be closed
// first. Callers must not use the terminal concurrently while NewSearch runs.
func NewSearch(t *Terminal) (*Search, error) {
	var ptr C.GhosttySearch
	if err := resultError(C.ghostty_search_new(nil, &ptr, t.ptr)); err != nil {
		return nil, err
	}
	return &Search{ptr: ptr}, nil
}

// Close releases resources owned by the search. If the terminal is still
// open, Close also releases state tracked by the terminal. Callers must not
// access the terminal concurrently while Close runs. The Search must not be
// used after Close returns.
func (s *Search) Close() {
	C.ghostty_search_free(s.ptr)
}

// Tick performs a bounded amount of search work and returns the new status.
// It uses only data already copied into the Search, so it may run while
// another goroutine accesses the terminal.
func (s *Search) Tick() (SearchStatus, error) {
	var status C.GhosttySearchStatus
	if err := resultError(C.ghostty_search_tick(s.ptr, &status)); err != nil {
		return 0, err
	}
	return SearchStatus(status), nil
}

// Feed copies recent terminal state into the Search. It updates matches in the
// active screen and viewport, then copies another portion of scrollback when
// needed. Each call performs a bounded amount of work. Call Feed periodically
// while the Search is in use so that later terminal changes are included.
//
// Callers must not access the terminal concurrently while Feed runs.
func (s *Search) Feed() error {
	return resultError(C.ghostty_search_feed(s.ptr))
}

// Run blocks until the Search has processed the terminal's current contents.
// It is convenient for one-time searches. Interactive applications should use
// [Search.Feed] and [Search.Tick] to control how much work happens at once.
//
// Callers must not access the terminal concurrently while Run executes.
func (s *Search) Run() error {
	return resultError(C.ghostty_search_run(s.ptr))
}

// GetMulti reads several low-level search fields in one library call. Each
// entry in values must point to storage of the type documented by the matching
// key. Keys that use a GhosttyString, GhosttySelection, or
// GhosttySelectionBuffer require storage matching the corresponding C type.
// Status and scroll policy fields require int32 storage because the underlying
// C enums are 32 bits.
// Prefer methods such as [Search.Status], [Search.MatchCount], and
// [Search.SelectedIndex] unless reducing calls into libghostty is important.
//
// On success, written equals len(keys). If a field cannot be read, written is
// the index of that field. Fields before that index have already been written.
//
// C: ghostty_search_get_multi
func (s *Search) GetMulti(keys []SearchData, values []unsafe.Pointer) (written int, err error) {
	if len(keys) != len(values) {
		return 0, errors.New("libghostty: keys and values must have the same length")
	}
	if len(keys) == 0 {
		return 0, nil
	}

	cKeys := make([]C.GhosttySearchData, len(keys))
	for i, key := range keys {
		cKeys[i] = C.GhosttySearchData(key)
	}
	// The Go output pointers are nested inside a C-allocated array, so pin them
	// explicitly for the duration of the call.
	var pinner runtime.Pinner
	for _, value := range values {
		if value != nil {
			pinner.Pin(value)
		}
	}
	defer pinner.Unpin()

	cValues, cValuesSize := cValuesArray(values)
	defer Free(unsafe.Pointer(cValues), cValuesSize)

	var cWritten C.size_t
	err = resultError(C.ghostty_search_get_multi(
		s.ptr,
		C.size_t(len(cKeys)),
		(*C.GhosttySearchData)(unsafe.Pointer(&cKeys[0])),
		cValues,
		&cWritten,
	))
	return int(cWritten), err
}

// SetNeedle sets the byte sequence to find. Matching is byte-for-byte except
// that ASCII letters are compared without regard to case. Search copies needle
// before SetNeedle returns.
//
// An empty needle clears the search and discards its results. Changing a
// non-empty needle restarts the search from the beginning. Setting an
// equivalent needle again preserves the current results. Equivalence uses the
// same ASCII case-insensitive comparison as matching.
func (s *Search) SetNeedle(needle []byte) error {
	if len(needle) == 0 {
		return s.set(C.GHOSTTY_SEARCH_OPT_NEEDLE, nil)
	}

	allocationSize := uintptr(len(needle))
	ptr := Alloc(allocationSize)
	if ptr == nil {
		return &Error{Result: ResultOutOfMemory}
	}
	defer Free(ptr, allocationSize)
	copy(unsafe.Slice((*byte)(ptr), len(needle)), needle)

	value := C.GhosttyString{
		ptr: (*C.uint8_t)(ptr),
		len: C.size_t(len(needle)),
	}
	return s.set(C.GHOSTTY_SEARCH_OPT_NEEDLE, unsafe.Pointer(&value))
}

// SelectNext first updates the Search from the terminal, then selects the next
// match toward older content. Selection wraps from the oldest match back to
// the newest match. It returns an error with result [ResultNoValue] when the
// search has no matches.
func (s *Search) SelectNext() error {
	return s.set(C.GHOSTTY_SEARCH_OPT_SELECT_NEXT, nil)
}

// SelectPrevious first updates the Search from the terminal, then selects the
// previous match toward newer content. Selection wraps from the newest match
// back to the oldest match. It returns an error with result [ResultNoValue]
// when the search has no matches.
func (s *Search) SelectPrevious() error {
	return s.set(C.GHOSTTY_SEARCH_OPT_SELECT_PREV, nil)
}

// SetScrollPolicy sets how selecting a match affects the terminal viewport.
// It only changes Search state and may run while another goroutine accesses
// the terminal.
func (s *Search) SetScrollPolicy(policy SearchScrollPolicy) error {
	value := C.GhosttySearchScroll(policy)
	return s.set(C.GHOSTTY_SEARCH_OPT_SELECT_SCROLL, unsafe.Pointer(&value))
}

// Status reports whether the Search can make progress without another call to
// [Search.Feed]. A complete Search may become stale after the terminal changes.
func (s *Search) Status() (SearchStatus, error) {
	var status C.GhosttySearchStatus
	if err := s.get(SearchDataStatus, unsafe.Pointer(&status)); err != nil {
		return 0, err
	}
	return SearchStatus(status), nil
}

// Needle returns a copy of the current search needle. It returns nil when the
// Search is idle. The returned bytes remain valid after later Search calls.
func (s *Search) Needle() ([]byte, error) {
	var value C.GhosttyString
	err := s.get(SearchDataNeedle, unsafe.Pointer(&value))
	if err != nil {
		if isResult(err, ResultNoValue) {
			return nil, nil
		}
		return nil, err
	}
	needle, ok := copyGhosttyString(value)
	if !ok {
		return nil, &Error{Result: ResultInvalidValue}
	}
	return needle, nil
}

// MatchCount returns the number of matches found so far in the active screen
// buffer, including its scrollback. The count is based on terminal contents
// copied by the most recent [Search.Feed] and can increase until
// [Search.Status] returns [SearchStatusComplete].
func (s *Search) MatchCount() (uint, error) {
	var count C.size_t
	if err := s.get(SearchDataMatchCount, unsafe.Pointer(&count)); err != nil {
		return 0, err
	}
	return uint(count), nil
}

// SelectedIndex returns the selected match's index in [Search.Matches]. It
// returns nil when no match is selected. Index zero is the newest match.
func (s *Search) SelectedIndex() (*uint, error) {
	var index C.size_t
	err := s.get(SearchDataSelectedIndex, unsafe.Pointer(&index))
	if err != nil {
		if isResult(err, ResultNoValue) {
			return nil, nil
		}
		return nil, err
	}
	value := uint(index)
	return &value, nil
}

// SelectedMatch returns the selected match as a Selection. It returns nil when
// no match is selected. The next operation that mutates the terminal may
// invalidate the Selection.
func (s *Search) SelectedMatch() (*Selection, error) {
	value := initCSelection()
	err := s.get(SearchDataSelectedMatch, unsafe.Pointer(&value))
	if err != nil {
		if isResult(err, ResultNoValue) {
			return nil, nil
		}
		return nil, err
	}
	match := selectionFromC(value)
	return &match, nil
}

// Matches returns the matches found so far on the active screen, including
// scrollback, from newest to oldest. The next operation that mutates the
// terminal may invalidate the returned selections.
func (s *Search) Matches() ([]Selection, error) {
	return s.selections(SearchDataMatches)
}

// ViewportMatches returns matches near the viewport as of the most recent
// [Search.Feed]. Applications can use these selections to draw search
// highlights. The result may include nearby off-screen matches. Use
// [Terminal.PointFromGridRef] with [PointTagViewport] to convert each endpoint,
// then ignore matches that fall outside the visible rows.
//
// The next operation that mutates the terminal may invalidate the returned
// selections.
func (s *Search) ViewportMatches() ([]Selection, error) {
	return s.selections(SearchDataViewportMatches)
}

// ScrollPolicy returns the policy used when a match is selected.
func (s *Search) ScrollPolicy() (SearchScrollPolicy, error) {
	var policy C.GhosttySearchScroll
	if err := s.get(SearchDataScrollPolicy, unsafe.Pointer(&policy)); err != nil {
		return 0, err
	}
	return SearchScrollPolicy(policy), nil
}

// selections reads either match list. The output buffer uses C memory because
// GhosttySelection contains C pointers.
func (s *Search) selections(data SearchData) ([]Selection, error) {
	var buffer C.GhosttySelectionBuffer
	err := s.get(data, unsafe.Pointer(&buffer))
	if err != nil && !isResult(err, ResultOutOfSpace) {
		return nil, err
	}

	capacity := int(buffer.len)
	if capacity == 0 {
		return nil, nil
	}

	allocationSize := uintptr(capacity) * C.sizeof_GhosttySelection
	ptr := Alloc(allocationSize)
	if ptr == nil {
		return nil, &Error{Result: ResultOutOfMemory}
	}
	defer Free(ptr, allocationSize)

	buffer.ptr = (*C.GhosttySelection)(ptr)
	buffer.cap = C.size_t(capacity)
	buffer.len = 0
	if err := s.get(data, unsafe.Pointer(&buffer)); err != nil {
		return nil, err
	}
	length := int(buffer.len)
	matches := make([]Selection, length)
	for i, value := range unsafe.Slice(buffer.ptr, length) {
		matches[i] = selectionFromC(value)
	}
	return matches, nil
}

// set writes one search option.
func (s *Search) set(option C.GhosttySearchOption, value unsafe.Pointer) error {
	return resultError(C.ghostty_search_set(s.ptr, option, value))
}

// get reads one search data field.
func (s *Search) get(data SearchData, value unsafe.Pointer) error {
	return resultError(C.ghostty_search_get(
		s.ptr,
		C.GhosttySearchData(data),
		value,
	))
}

// isResult reports whether err is a libghostty Error with the requested
// result code.
func isResult(err error, result Result) bool {
	var ghosttyErr *Error
	return errors.As(err, &ghosttyErr) && ghosttyErr.Result == result
}

package libghostty

/*
#include <ghostty/vt.h>
*/
import "C"

// The version of the libghostty-vt headers this package was compiled against.
// These are fixed at compile time. [Version] and [GetBuildInfo] report the
// library that is actually loaded, which can differ when linking the shared
// library.
//
// A program works with any library of the same major version that is at least
// as new as the headers. An older library may be missing functions that newer
// headers declare.
//
// These always hold the number of the release, even in a development build.
// For example, a development build before 1.1.0 reports "1.1.0" here while
// [Version] returns the full "1.1.0-dev+abc1234".
const (
	// HeaderVersionMajor is the major version of the headers.
	// C: GHOSTTY_VT_VERSION_MAJOR
	HeaderVersionMajor = C.GHOSTTY_VT_VERSION_MAJOR

	// HeaderVersionMinor is the minor version of the headers.
	// C: GHOSTTY_VT_VERSION_MINOR
	HeaderVersionMinor = C.GHOSTTY_VT_VERSION_MINOR

	// HeaderVersionPatch is the patch version of the headers.
	// C: GHOSTTY_VT_VERSION_PATCH
	HeaderVersionPatch = C.GHOSTTY_VT_VERSION_PATCH

	// HeaderVersion is the version of the headers as a string, such as "1.0.0".
	// It never has a pre-release or build suffix such as "-dev".
	// C: GHOSTTY_VT_VERSION
	HeaderVersion = C.GHOSTTY_VT_VERSION
)

package libghostty

/*
#include <ghostty/vt.h>
#include <ghostty/vt/build_info.h>
*/
import "C"

import "unsafe"

// OptimizeMode identifies the optimization mode the library was built with.
// C: GhosttyOptimizeMode
type OptimizeMode int

const (
	// OptimizeDebug is the debug optimization mode.
	OptimizeDebug OptimizeMode = C.GHOSTTY_OPTIMIZE_DEBUG

	// OptimizeReleaseSafe is the release-safe optimization mode.
	OptimizeReleaseSafe OptimizeMode = C.GHOSTTY_OPTIMIZE_RELEASE_SAFE

	// OptimizeReleaseSmall is the release-small optimization mode.
	OptimizeReleaseSmall OptimizeMode = C.GHOSTTY_OPTIMIZE_RELEASE_SMALL

	// OptimizeReleaseFast is the release-fast optimization mode.
	OptimizeReleaseFast OptimizeMode = C.GHOSTTY_OPTIMIZE_RELEASE_FAST
)

// BuildInfo holds compile-time build configuration of libghostty-vt.
// All values are constant for the lifetime of the process.
// C: GhosttyBuildInfo (query enum)
type BuildInfo struct {
	// SIMD reports whether SIMD-accelerated code paths are enabled.
	SIMD bool

	// KittyGraphics reports whether Kitty graphics protocol support
	// is available.
	KittyGraphics bool

	// TmuxControlMode reports whether tmux control mode support
	// is available.
	TmuxControlMode bool

	// Optimize is the optimization mode the library was built with.
	Optimize OptimizeMode

	// VersionString is the full version string
	// (e.g. "1.2.3" or "1.2.3-dev+abcdef").
	VersionString string

	// VersionMajor is the major version number.
	VersionMajor uint

	// VersionMinor is the minor version number.
	VersionMinor uint

	// VersionPatch is the patch version number.
	VersionPatch uint

	// VersionPre is the pre-release metadata string (e.g. "alpha",
	// "beta", or "dev"). Empty if no pre-release metadata is present.
	VersionPre string

	// VersionBuild is the build metadata string (e.g. commit hash).
	// Empty if no build metadata is present.
	VersionBuild string

	// FeatureSnapshot reports whether terminal snapshots are available.
	FeatureSnapshot bool

	// FeatureFormatter reports whether the formatter is available.
	FeatureFormatter bool

	// FeatureSelection reports whether selection APIs are available.
	FeatureSelection bool

	// FeatureSearch reports whether terminal search is available.
	FeatureSearch bool

	// FeatureRenderState reports whether the render state is available.
	FeatureRenderState bool

	// FeatureInputEncode reports whether key, mouse, focus, and paste
	// encoding are available.
	FeatureInputEncode bool

	// FeatureColor reports whether color utilities are available.
	FeatureColor bool

	// FeatureGridIntrospection reports whether grid references and the
	// cell and row accessors are available.
	FeatureGridIntrospection bool

	// FeatureGlyphProtocol reports whether applications can register
	// custom glyphs with the Glyph Protocol.
	FeatureGlyphProtocol bool
}

// GetBuildInfo queries all compile-time build configuration values
// and returns them in a single BuildInfo struct.
func GetBuildInfo() (BuildInfo, error) {
	var info BuildInfo

	var simd C.bool
	if err := resultError(C.ghostty_build_info(C.GHOSTTY_BUILD_INFO_SIMD, unsafe.Pointer(&simd))); err != nil {
		return info, err
	}
	info.SIMD = bool(simd)

	var kitty C.bool
	if err := resultError(C.ghostty_build_info(C.GHOSTTY_BUILD_INFO_KITTY_GRAPHICS, unsafe.Pointer(&kitty))); err != nil {
		return info, err
	}
	info.KittyGraphics = bool(kitty)

	var tmux C.bool
	if err := resultError(C.ghostty_build_info(C.GHOSTTY_BUILD_INFO_TMUX_CONTROL_MODE, unsafe.Pointer(&tmux))); err != nil {
		return info, err
	}
	info.TmuxControlMode = bool(tmux)

	var opt C.GhosttyOptimizeMode
	if err := resultError(C.ghostty_build_info(C.GHOSTTY_BUILD_INFO_OPTIMIZE, unsafe.Pointer(&opt))); err != nil {
		return info, err
	}
	info.Optimize = OptimizeMode(opt)

	var verStr C.GhosttyString
	if err := resultError(C.ghostty_build_info(C.GHOSTTY_BUILD_INFO_VERSION_STRING, unsafe.Pointer(&verStr))); err != nil {
		return info, err
	}
	info.VersionString = C.GoStringN((*C.char)(unsafe.Pointer(verStr.ptr)), C.int(verStr.len))

	var major C.size_t
	if err := resultError(C.ghostty_build_info(C.GHOSTTY_BUILD_INFO_VERSION_MAJOR, unsafe.Pointer(&major))); err != nil {
		return info, err
	}
	info.VersionMajor = uint(major)

	var minor C.size_t
	if err := resultError(C.ghostty_build_info(C.GHOSTTY_BUILD_INFO_VERSION_MINOR, unsafe.Pointer(&minor))); err != nil {
		return info, err
	}
	info.VersionMinor = uint(minor)

	var patch C.size_t
	if err := resultError(C.ghostty_build_info(C.GHOSTTY_BUILD_INFO_VERSION_PATCH, unsafe.Pointer(&patch))); err != nil {
		return info, err
	}
	info.VersionPatch = uint(patch)

	var verPre C.GhosttyString
	if err := resultError(C.ghostty_build_info(C.GHOSTTY_BUILD_INFO_VERSION_PRE, unsafe.Pointer(&verPre))); err != nil {
		return info, err
	}
	info.VersionPre = C.GoStringN((*C.char)(unsafe.Pointer(verPre.ptr)), C.int(verPre.len))

	var verBuild C.GhosttyString
	if err := resultError(C.ghostty_build_info(C.GHOSTTY_BUILD_INFO_VERSION_BUILD, unsafe.Pointer(&verBuild))); err != nil {
		return info, err
	}
	info.VersionBuild = C.GoStringN((*C.char)(unsafe.Pointer(verBuild.ptr)), C.int(verBuild.len))

	// Each optional feature has a key that reports whether the library was
	// built with it. A call into a missing feature returns [ErrUnsupported].
	features := [...]struct {
		key C.GhosttyBuildInfo
		out *bool
	}{
		{C.GHOSTTY_BUILD_INFO_FEATURE_SNAPSHOT, &info.FeatureSnapshot},
		{C.GHOSTTY_BUILD_INFO_FEATURE_FORMATTER, &info.FeatureFormatter},
		{C.GHOSTTY_BUILD_INFO_FEATURE_SELECTION, &info.FeatureSelection},
		{C.GHOSTTY_BUILD_INFO_FEATURE_SEARCH, &info.FeatureSearch},
		{C.GHOSTTY_BUILD_INFO_FEATURE_RENDER_STATE, &info.FeatureRenderState},
		{C.GHOSTTY_BUILD_INFO_FEATURE_INPUT_ENCODE, &info.FeatureInputEncode},
		{C.GHOSTTY_BUILD_INFO_FEATURE_COLOR, &info.FeatureColor},
		{C.GHOSTTY_BUILD_INFO_FEATURE_GRID_INTROSPECTION, &info.FeatureGridIntrospection},
		{C.GHOSTTY_BUILD_INFO_FEATURE_GLYPH_PROTOCOL, &info.FeatureGlyphProtocol},
	}
	for _, feature := range features {
		// A shared library older than the headers doesn't know a newer key,
		// which means it doesn't have that feature, so that isn't a failure.
		var enabled C.bool
		result := C.ghostty_build_info(feature.key, unsafe.Pointer(&enabled))
		if result == C.GHOSTTY_UNSUPPORTED {
			continue
		}
		if err := resultError(result); err != nil {
			return info, err
		}
		*feature.out = bool(enabled)
	}

	return info, nil
}

// Version returns the full version of the loaded library, such as
// "1.1.0-dev+9a018f5cc". It is the same text as [BuildInfo.VersionString], but
// needs no query and cannot fail.
//
// Compare it with [HeaderVersion] to see what this package was compiled
// against. The two can differ when linking the shared library, which is where
// the check matters: every 1.x release uses the same shared library name, so
// the system can load an older 1.x library than the one this package was built
// against.
// C: ghostty_version
func Version() string {
	return C.GoString(C.ghostty_version())
}

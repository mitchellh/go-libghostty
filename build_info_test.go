package libghostty

import (
	"fmt"
	"strings"
	"testing"
)

func TestGetBuildInfo(t *testing.T) {
	info, err := GetBuildInfo()
	if err != nil {
		t.Fatal(err)
	}

	// Version string should be non-empty.
	if info.VersionString == "" {
		t.Fatal("expected non-empty version string")
	}

	// At least one version component should be non-zero.
	if info.VersionMajor == 0 && info.VersionMinor == 0 && info.VersionPatch == 0 {
		t.Fatal("expected at least one non-zero version component")
	}

	// Optimize mode should be a known value.
	switch info.Optimize {
	case OptimizeDebug, OptimizeReleaseSafe, OptimizeReleaseSmall, OptimizeReleaseFast:
	default:
		t.Fatalf("unexpected optimize mode: %d", info.Optimize)
	}
}

// TestVersion checks that every way of reading the version agrees. The headers
// and the library always come from the same commit in this build, so a mismatch
// here means a version lookup is wired up wrong.
func TestVersion(t *testing.T) {
	info, err := GetBuildInfo()
	if err != nil {
		t.Fatal(err)
	}

	if got := Version(); got != info.VersionString {
		t.Fatalf("Version() = %q, BuildInfo = %q", got, info.VersionString)
	}

	want := fmt.Sprintf("%d.%d.%d", HeaderVersionMajor, HeaderVersionMinor, HeaderVersionPatch)
	if HeaderVersion != want {
		t.Fatalf("HeaderVersion = %q, want %q", HeaderVersion, want)
	}
	if !strings.HasPrefix(Version(), HeaderVersion) {
		t.Fatalf("library %q doesn't match headers %q", Version(), HeaderVersion)
	}
}

// TestBuildInfoFeatures checks the feature keys. The library this package builds
// against has every feature, and the rest of the tests rely on them.
func TestBuildInfoFeatures(t *testing.T) {
	info, err := GetBuildInfo()
	if err != nil {
		t.Fatal(err)
	}

	for name, enabled := range map[string]bool{
		"snapshot":           info.FeatureSnapshot,
		"formatter":          info.FeatureFormatter,
		"selection":          info.FeatureSelection,
		"search":             info.FeatureSearch,
		"render state":       info.FeatureRenderState,
		"input encode":       info.FeatureInputEncode,
		"color":              info.FeatureColor,
		"grid introspection": info.FeatureGridIntrospection,
		"glyph protocol":     info.FeatureGlyphProtocol,
	} {
		if !enabled {
			t.Errorf("expected feature %q to be enabled", name)
		}
	}
}

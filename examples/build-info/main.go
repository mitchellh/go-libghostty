// Example build-info demonstrates querying libghostty's compile-time
// build configuration using the GetBuildInfo API.
package main

import (
	"fmt"
	"log"

	"go.mitchellh.com/libghostty"
)

func boolStr(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}

func main() {
	info, err := libghostty.GetBuildInfo()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("SIMD: %s\n", boolStr(info.SIMD))
	fmt.Printf("Kitty graphics: %s\n", boolStr(info.KittyGraphics))
	fmt.Printf("Tmux control mode: %s\n", boolStr(info.TmuxControlMode))

	fmt.Printf("Search: %s\n", boolStr(info.FeatureSearch))
	fmt.Printf("Snapshots: %s\n", boolStr(info.FeatureSnapshot))

	// The headers are fixed when this program is compiled, while the library
	// can differ when it is linked as a shared library. The major versions must
	// match, and the library may be newer than the headers but not older.
	fmt.Printf("Headers: %s\n", libghostty.HeaderVersion)
	fmt.Printf("Library: %s\n", libghostty.Version())
	if info.VersionMajor != libghostty.HeaderVersionMajor ||
		info.VersionMinor < libghostty.HeaderVersionMinor {
		log.Fatal("library is not compatible with the headers")
	}

	fmt.Printf("Version: %s\n", info.VersionString)
	fmt.Printf("Version major: %d\n", info.VersionMajor)
	fmt.Printf("Version minor: %d\n", info.VersionMinor)
	fmt.Printf("Version patch: %d\n", info.VersionPatch)
	if info.VersionBuild != "" {
		fmt.Printf("Version build: %s\n", info.VersionBuild)
	} else {
		fmt.Printf("Version build: (none)\n")
	}
}

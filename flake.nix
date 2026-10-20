{
  description = "go-libghostty";

  inputs = {
    nixpkgs.url = "https://channels.nixos.org/nixpkgs-unstable/nixexprs.tar.xz";
    flake-utils.url = "github:numtide/flake-utils";
    zig = {
      url = "github:mitchellh/zig-overlay";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    nixpkgs,
    flake-utils,
    zig,
    ...
  }:
    flake-utils.lib.eachDefaultSystem (
      system: let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        devShells.default = pkgs.mkShell {
          packages = [
            pkgs.cmake
            pkgs.go
            pkgs.pinact
            pkgs.pkg-config
            zig.packages.${system}."0.17.0"
          ];

          shellHook =
            ''
              export PKG_CONFIG_PATH="$PWD/build/_deps/ghostty-src/zig-out/share/pkgconfig''${PKG_CONFIG_PATH:+:$PKG_CONFIG_PATH}"
              export DYLD_LIBRARY_PATH="$PWD/build/_deps/ghostty-src/zig-out/lib''${DYLD_LIBRARY_PATH:+:$DYLD_LIBRARY_PATH}"
              export LD_LIBRARY_PATH="$PWD/build/_deps/ghostty-src/zig-out/lib''${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
            ''
            + pkgs.lib.optionalString pkgs.stdenv.hostPlatform.isDarwin ''
              # The libghostty-vt build post-processes its static archive
              # with the system Xcode tools (`xcrun nmedit`, see upstream
              # src/build/libsystem_override.sh). Nix's SDK env vars and
              # xcbuild xcrun shim can't find those tools, so drop them and
              # rely on the system installation, matching upstream's shell.
              unset SDKROOT
              unset DEVELOPER_DIR
              export PATH=$(echo "$PATH" | awk -v RS=: -v ORS=: '$0 !~ /xcrun/ || $0 == "/usr/bin" {print}' | sed 's/:$//')
            '';
        };
      }
    );
}

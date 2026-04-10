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
            zig.packages.${system}."0.15.2"
          ] ++ pkgs.lib.optionals pkgs.stdenv.hostPlatform.isLinux [
            pkgs.libcxx
          ];

          shellHook = ''
            export PKG_CONFIG_PATH="$PWD/build/_deps/ghostty-src/zig-out/share/pkgconfig''${PKG_CONFIG_PATH:+:$PKG_CONFIG_PATH}"
            export DYLD_LIBRARY_PATH="$PWD/build/_deps/ghostty-src/zig-out/lib''${DYLD_LIBRARY_PATH:+:$DYLD_LIBRARY_PATH}"
            export LD_LIBRARY_PATH="$PWD/build/_deps/ghostty-src/zig-out/lib''${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
          '';
        };
      }
    );
}

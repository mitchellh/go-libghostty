# Go Libghostty Bindings

Go bindings for `libghostty-vt`.

This project uses [cgo](https://pkg.go.dev/cmd/cgo) but `libghostty-vt`
only depends on libc, so it is very easy to static link and very
easy to cross-compile. The bindings default to static linking for this
reason.

> [!WARNING]
>
> **I'm not promising any API stability yet.** This is a new project and the
> API may change as necessary. The underlying functionality is very stable,
> but the Go API is still being designed.

## Example

```go
package main

import (
 "fmt"
 "log"

 "go.mitchellh.com/libghostty"
)

func main() {
 term, err := libghostty.NewTerminal(libghostty.WithSize(80, 24))
 if err != nil {
  log.Fatal(err)
 }
 defer term.Close()

 // Feed VT data — bold green "world", then plain text.
 fmt.Fprintf(term, "Hello, \033[1;32mworld\033[0m!\r\n")

 // Format the terminal contents as plain text.
 f, err := libghostty.NewFormatter(term,
  libghostty.WithFormatterFormat(libghostty.FormatterFormatPlain),
  libghostty.WithFormatterTrim(true),
 )
 if err != nil {
  log.Fatal(err)
 }
 defer f.Close()

 output, _ := f.FormatString()
 fmt.Println(output) // Hello, world!
}
```

More examples are in the [`examples/`](examples/) directory.

## Usage

Add the module to your Go project:

```shell
go get go.mitchellh.com/libghostty
```

This is a cgo package that links `libghostty-vt` via `pkg-config`. By
default it links statically. Before building your project, you need the
library installed. Either install it system-wide or set `PKG_CONFIG_PATH`
to point to a local checkout:

```shell
export PKG_CONFIG_PATH=/path/to/libghostty-vt/share/pkgconfig
```

To link dynamically instead (requires the shared library at runtime,
so you'll also need to set the library path):

```shell
go build -tags dynamic
```

See the [Ghostty docs](https://ghostty.org/docs/install/build) for
building `libghostty-vt` from source.

### Cross-Compilation

Because `libghostty-vt` only depends on libc, cross-compilation is
straightforward using [Zig](https://ziglang.org/) as the C compiler.
Zig is already required to build `libghostty-vt`, so no extra tooling
is needed. You don't need to write any Zig code, we're just using
Zig as a C/C++ compiler.

First, build `libghostty-vt` for your target (from the ghostty source tree):

```shell
zig build -Demit-lib-vt -Dtarget=x86_64-linux-gnu --prefix /tmp/ghostty-linux-amd64
```

Then cross-compile your Go project with `zig cc`:

```shell
CGO_ENABLED=1 \
GOOS=linux GOARCH=amd64 \
CC="zig cc -target x86_64-linux-gnu" \
CXX="zig c++ -target x86_64-linux-gnu" \
CGO_CFLAGS="-I/tmp/ghostty-linux-amd64/include -DGHOSTTY_STATIC" \
CGO_LDFLAGS="-L/tmp/ghostty-linux-amd64/lib -lghostty-vt" \
go build ./...
```

Supported targets include `x86_64-linux-gnu`, `aarch64-linux-gnu`,
`x86_64-macos`, `aarch64-macos`, `x86_64-windows-gnu`, and
`aarch64-windows-gnu`.

If you are using ghostty's CMake integration via `FetchContent`, the
`ghostty_vt_add_target()` function handles the zig build for you:

```cmake
FetchContent_MakeAvailable(ghostty)
ghostty_vt_add_target(NAME linux-amd64 ZIG_TARGET x86_64-linux-gnu)
```

See the [ghostty CMakeLists.txt](https://github.com/ghostty-org/ghostty/blob/main/CMakeLists.txt)
for full documentation of `ghostty_vt_add_target()`.

## Development

CMake fetches and builds `libghostty-vt` automatically. CMake is only
required and used for development of this module. For actual downstream
usage, you can get `libghostty-vt` available however you like (e.g. system
package, local checkout, etc.).

You need [Zig](https://ghostty.org/docs/install/build) and CMake on your PATH.

```shell
make build
make test

# If in a Nix dev shell:
go build
go test
```

If you use the Nix dev shell (`nix develop`), `go build` and `go test`
work directly — the shell configures all paths automatically.

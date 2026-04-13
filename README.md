# Go Libghostty Bindings

Go bindings for `libghostty-vt`.

This project uses [cgo](https://pkg.go.dev/cmd/cgo) but `libghostty-vt`
only depends on libc/libc++, so it is very easy to static link and very
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

 "github.com/mitchellh/go-libghostty"
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
go get github.com/mitchellh/go-libghostty
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

## Development

CMake fetches and builds `libghostty-vt` automatically. CMake is only
required and used for development of this module. For actual downstream
usage, you can get `libghostty-vt` available however you like (e.g. system
package, local checkout, etc.).

You need [Zig](https://ghostty.org/docs/install/build), `pkg-config`, and CMake on your PATH.

```shell
make build
make test

# If in a Nix dev shell:
go build
go test
```

If you use the Nix dev shell (`nix develop`), `go build` and `go test`
work directly — the shell configures all paths automatically.

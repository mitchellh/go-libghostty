# Missing APIs

## WASM-only Not Bound

These APIs are guarded by `#ifdef __wasm__` upstream and are not available in the native cgo build.

- [ ] WebAssembly allocation helpers (`wasm.h`)
  - `ghostty_wasm_alloc()`
  - `ghostty_wasm_free()`
  - `ghostty_wasm_alloc_opaque()`
  - `ghostty_wasm_free_opaque()`
  - `ghostty_wasm_take_opaque()`

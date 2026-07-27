# Missing APIs

## WASM-only Not Bound

These APIs are guarded by `#ifdef __wasm__` upstream and are not available in the native cgo build.

- [ ] WebAssembly allocation helpers (`wasm.h`)
  - `ghostty_wasm_alloc_opaque()`
  - `ghostty_wasm_free_opaque()`
  - `ghostty_wasm_alloc_u8_array()`
  - `ghostty_wasm_free_u8_array()`
  - `ghostty_wasm_alloc_u16_array()`
  - `ghostty_wasm_free_u16_array()`
  - `ghostty_wasm_alloc_u8()`
  - `ghostty_wasm_free_u8()`
  - `ghostty_wasm_alloc_usize()`
  - `ghostty_wasm_free_usize()`
  - `ghostty_wasm_alloc_sgr_attribute()` (`sgr.h`)
  - `ghostty_wasm_free_sgr_attribute()` (`sgr.h`)

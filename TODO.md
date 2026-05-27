# Missing APIs

## Not Bound

- [x] Key encoding (`key.h`, `key/encoder.h`, `key/event.h`)
- [x] Mouse encoding (`mouse.h`, `mouse/encoder.h`, `mouse/event.h`)
- [ ] OSC parser (`osc.h`)
  - `ghostty_osc_new()`
  - `ghostty_osc_free()`
  - `ghostty_osc_reset()`
  - `ghostty_osc_next()`
  - `ghostty_osc_end()`
  - `ghostty_osc_command_type()`
  - `ghostty_osc_command_data()`
- [ ] SGR parser (`sgr.h`)
  - `ghostty_sgr_new()`
  - `ghostty_sgr_free()`
  - `ghostty_sgr_reset()`
  - `ghostty_sgr_set_params()`
  - `ghostty_sgr_next()`
  - `ghostty_sgr_unknown_full()`
  - `ghostty_sgr_unknown_partial()`
  - `ghostty_sgr_attribute_tag()`
  - `ghostty_sgr_attribute_value()`
- [x] Paste utilities (`paste.h`)
- [x] Focus encoding (`focus.h`)
- [x] Kitty graphics (`kitty_graphics.h`)
- [x] Allocator (`allocator.h` — `ghostty_alloc`, `ghostty_free`)
- [x] Terminal selection helpers (`selection.h`)
- [x] Selection gesture APIs (`selection.h`)
- [x] Render-state row selection (`GHOSTTY_RENDER_STATE_ROW_DATA_SELECTION`)
- [x] Build info (`build_info.h`)

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

## Partially Bound

- [ ] `ghostty_mode_report_encode()`
- [ ] `ghostty_size_report_encode()`
- [ ] `ghostty_type_json()`
- [ ] `ghostty_style_default()`
- [ ] `ghostty_color_rgb_get()`
- [ ] `ghostty_formatter_format_buf()`

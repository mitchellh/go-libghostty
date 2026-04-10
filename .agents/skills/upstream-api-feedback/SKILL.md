---
name: upstream-api-feedback
description: >-
  Reviews Go bindings against upstream libghostty-vt C headers and
  produces concrete upstream API improvement suggestions. Use when
  asked to review bindings for upstream feedback, suggest C API
  changes, find API friction, or improve the upstream API.
---

# Upstream API Feedback

Analyzes the Go bindings in this project against the upstream
libghostty-vt C headers to identify concrete, actionable API
changes that would make the bindings more performant, idiomatic,
or easier to use. Output is written to directly to the chat.

## When to use

- The user asks for upstream API feedback or suggestions.
- The user asks to review bindings for performance or ergonomics.
- After binding a new API, to check for friction points.

## Workflow

### 1. Identify scope

Determine which bindings to review:

- **Specific file** (e.g. "review kitty_graphics.go"): Focus on
  that file and its corresponding C header.
- **All bindings** (e.g. "review everything"): Scan all `*.go`
  files (excluding `_test.go`) and cross-reference with headers.
- **Recent changes**: Use `jj diff` or `git diff` to find what
  changed and focus on those files.

### 2. Read the Go bindings and C headers together

For each file in scope:

1. Read the Go file to understand the binding patterns used.
2. Read the corresponding C header in
   `build/_deps/ghostty-src/zig-out/include/ghostty/vt/`.
3. Note every CGo crossing — each `C.ghostty_*()` call.

### 3. Apply the analysis checklist

For each API surface, check for these patterns:

#### Multi-call data fetching (highest impact)

Look for Go functions that make multiple `C.ghostty_*_get()` calls
to fetch logically related fields from the same object. Each CGo
call has ~100ns overhead that compounds in hot paths.

**Signal**: A Go type whose methods each call the same C
`_get(handle, ENUM_VARIANT, &out)` function with different enum
values — especially when callers typically need several fields
together.

**Suggestion**: A sized struct (like `GhosttyRenderStateColors`)
that returns all fields in a single call. Reference the existing
`GHOSTTY_INIT_SIZED` pattern.

#### Pointer/length splits

Look for cases where pointer and length are separate enum variants
in the same `_get()` API (e.g. `DATA_PTR` + `DATA_LEN`). These
are semantically one value split across two calls.

**Signal**: The Go binding must make two sequential calls and has
no atomicity guarantee between them.

**Suggestion**: Either fold into the sized struct above, use
`GhosttyString`-style `{ptr, len}` as a single variant, or add a
dedicated function.

#### Repeated parameter triples

Look for multiple C functions that take the same parameter
combination (e.g. `(iterator, image, terminal)`) and are typically
called together per iteration step.

**Signal**: The Go code calls 3-4 functions with identical
arguments in sequence during a loop body.

**Suggestion**: A combined function returning a struct with all
results, cutting N CGo crossings to 1.

#### Two-phase initialization

Look for patterns where an object must be allocated, then populated
via a separate call before it's usable (e.g. `_new()` then
`_get(POPULATE, &handle)`).

**Signal**: The Go binding wraps this in a helper but the C API
still requires two calls where one would suffice.

**Suggestion**: A combined constructor, or making the populate step
part of `_new()`.

#### Missing convenience variants

Look for C APIs that could offer a simpler overload for the common
case while keeping the flexible version.

**Signal**: The Go binding wraps a complex C API with a simpler Go
function that most callers use, hiding parameters that are almost
always the same value.

**Suggestion**: A C convenience function for the common case.

### 4. Assess impact

Rate each suggestion:

- **Hot path**: Called per-frame or per-placement in a render loop.
  CGo overhead is multiplied by iteration count. High impact.
- **Setup path**: Called once during initialization or
  configuration. Low impact — ergonomics matter more than perf.
- **Ergonomic**: Doesn't affect performance but makes the API
  harder to use correctly or requires awkward binding code.

### 5. Write the output

Write findings to chat. Format:

```markdown
# Upstream API Feedback

Summary of the review scope and methodology.

## Suggestion Title

**Impact**: Hot path / Setup path / Ergonomic
**Files**: `header.h`, `binding.go`

Description of the current pattern, why it's suboptimal, and the
concrete C API change. Include a struct/function signature sketch.

## Next Suggestion

...
```

Group suggestions by impact (hot path first). Include concrete C
type/function signatures — not vague ideas. Reference existing
libghostty-vt patterns (like `GhosttyRenderStateColors`) as
precedent when applicable.

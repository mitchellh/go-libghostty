cmake_minimum_required(VERSION 3.19)
project(go-libghostty LANGUAGES C)

include(FetchContent)
FetchContent_Declare(ghostty
    GIT_REPOSITORY https://github.com/ghostty-org/ghostty.git
    GIT_TAG 2ed382a15566b267c32fae440b065f7844b15bfb
)
FetchContent_MakeAvailable(ghostty)

# Optional cross-compilation targets.
#
# By default this CMakeLists only builds the native libghostty-vt so
# that local `make build` / `make test` cycles stay fast. CI opts in
# to specific cross targets by passing this variable, which produces
# the corresponding `ghostty-vt-static-<NAME>` IMPORTED target with
# outputs under `build/ghostty-<NAME>/`.
#
# The value is a semicolon-separated list of `<name>=<zig-target>`
# entries, e.g.:
#
#   cmake -B build \
#     -DGHOSTTY_VT_CROSS_TARGETS="linux-amd64=x86_64-linux-gnu"
#
# Multiple targets can be configured at once by separating with `;`.
set(GHOSTTY_VT_CROSS_TARGETS "" CACHE STRING
    "Semicolon-separated <name>=<zig-target> cross targets to build")

foreach(_entry IN LISTS GHOSTTY_VT_CROSS_TARGETS)
    if(NOT _entry MATCHES "^([^=]+)=(.+)$")
        message(FATAL_ERROR
            "GHOSTTY_VT_CROSS_TARGETS entry must be <name>=<zig-target>: ${_entry}")
    endif()
    ghostty_vt_add_target(NAME ${CMAKE_MATCH_1} ZIG_TARGET ${CMAKE_MATCH_2})
endforeach()

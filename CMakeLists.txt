cmake_minimum_required(VERSION 3.19)
project(go-libghostty LANGUAGES C)

include(FetchContent)
FetchContent_Declare(ghostty
    GIT_REPOSITORY https://github.com/ghostty-org/ghostty.git
    GIT_TAG c36b458ad57a95869745e405c9d8d45104a97773
    PATCH_COMMAND git apply "${CMAKE_CURRENT_LIST_DIR}/disable-xcframework.patch"
)
FetchContent_MakeAvailable(ghostty)

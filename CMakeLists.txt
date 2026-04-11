cmake_minimum_required(VERSION 3.19)
project(go-libghostty LANGUAGES C)

include(FetchContent)
FetchContent_Declare(ghostty
    GIT_REPOSITORY https://github.com/ghostty-org/ghostty.git
    GIT_TAG c36b458ad57a95869745e405c9d8d45104a97773
)
FetchContent_MakeAvailable(ghostty)

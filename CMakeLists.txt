cmake_minimum_required(VERSION 3.19)
project(go-libghostty LANGUAGES C)

include(FetchContent)
FetchContent_Declare(ghostty
    GIT_REPOSITORY https://github.com/ghostty-org/ghostty.git
    GIT_TAG 7421b4b13f87e101d4bbcedd4da84886ceae4e7b
)
FetchContent_MakeAvailable(ghostty)

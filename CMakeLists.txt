cmake_minimum_required(VERSION 3.19)
project(go-libghostty LANGUAGES C)

include(FetchContent)
FetchContent_Declare(ghostty
    GIT_REPOSITORY https://github.com/ghostty-org/ghostty.git
    GIT_TAG 43a05dc968eda9bfa2196d66ba1819daf510b62a
)
FetchContent_MakeAvailable(ghostty)

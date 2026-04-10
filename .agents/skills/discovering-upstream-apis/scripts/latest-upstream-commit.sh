#!/usr/bin/env bash
# Prints the latest commit SHA on the main branch of the upstream
# ghostty repo. Compares it against the pinned commit in
# CMakeLists.txt and reports whether an update is needed.

set -euo pipefail

REPO="ghostty-org/ghostty"
BRANCH="main"
CMAKE="CMakeLists.txt"

# Get the currently pinned commit from CMakeLists.txt.
pinned=$(grep -oP '(?<=GIT_TAG )[0-9a-f]+' "$CMAKE" 2>/dev/null || true)
if [ -z "$pinned" ]; then
    echo "ERROR: could not find GIT_TAG in $CMAKE" >&2
    exit 1
fi

# Fetch the latest commit from GitHub.
latest=$(git ls-remote "https://github.com/${REPO}.git" "refs/heads/${BRANCH}" \
    | awk '{print $1}')
if [ -z "$latest" ]; then
    echo "ERROR: could not fetch latest commit from ${REPO}" >&2
    exit 1
fi

echo "pinned:  $pinned"
echo "latest:  $latest"

if [ "$pinned" = "$latest" ]; then
    echo "status:  up-to-date"
else
    echo "status:  update-available"
fi

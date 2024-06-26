#!/bin/sh

# set -eou pipefail

git config --global --add safe.directory /github/workspace

# Create Pull Request (by default on current branch)
gh pr create \
    --title "${INPUT_TITLE_PREFIX} ${VERSION}" \
    --body "This is an autogenerated PR to prepare for the release" \
    --reviewer "${INPUT_REVIEWERS}"
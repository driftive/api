#!/bin/bash
set -e

# Parse arguments
DRY_RUN=false
if [ "$1" == "--dry-run" ]; then
    DRY_RUN=true
fi

# Check if on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo "Error: Not on main branch (currently on: $CURRENT_BRANCH)"
    echo "Please switch to main branch before creating a release tag"
    exit 1
fi
echo "✓ On main branch"

# Fetch all tags and latest changes from remote
echo "Fetching tags and latest changes..."
git fetch --tags
git fetch origin main

# Check if local main is up to date with origin/main
LOCAL=$(git rev-parse main)
REMOTE=$(git rev-parse origin/main)

if [ "$LOCAL" != "$REMOTE" ]; then
    echo "Error: Local main branch is not up to date with origin/main"
    echo "Please pull the latest changes: git pull origin main"
    exit 1
fi
echo "✓ Branch is up to date with origin/main"

# Get the latest stable tag (excluding beta versions)
# Sort by version and get the latest
LATEST_STABLE_TAG=$(git tag -l | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -1)

if [ -z "$LATEST_STABLE_TAG" ]; then
    echo "No stable tags found, starting with v0.1.0"
    NEW_VERSION="v0.1.0"
else
    echo "Latest stable tag: $LATEST_STABLE_TAG"

    # Parse version components (remove 'v' prefix)
    VERSION=${LATEST_STABLE_TAG#v}
    IFS='.' read -r MAJOR MINOR PATCH <<< "$VERSION"

    # Bump minor version and reset patch to 0
    MINOR=$((MINOR + 1))
    NEW_VERSION="v${MAJOR}.${MINOR}.0"
    echo "Bumping minor version: $LATEST_STABLE_TAG -> $NEW_VERSION"
fi

echo ""
echo "New version: $NEW_VERSION"

if [ "$DRY_RUN" = true ]; then
    echo ""
    echo "[DRY RUN] Would create and push tag: $NEW_VERSION"
    echo "[DRY RUN] Run without --dry-run to create the tag"
else
    echo ""
    echo "Creating tag: $NEW_VERSION"
    git tag "$NEW_VERSION"

    echo "Pushing tag to origin..."
    git push origin "$NEW_VERSION"

    echo ""
    echo "✓ Successfully created and pushed tag: $NEW_VERSION"
fi

#!/usr/bin/env bash

set -e

release_lib_validate_environment() {
  if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "Error: Not in a git repository"
    return 1
  fi

  if [ -n "$(git status --porcelain --untracked-files=no)" ]; then
    echo "Error: Working directory has uncommitted changes"
    return 1
  fi

  local current_branch
  current_branch=$(git branch --show-current)
  if [ "$current_branch" != "main" ]; then
    echo "Error: Not on main branch (current: $current_branch)"
    return 1
  fi

  return 0
}

release_lib_validate_changelog() {
  if [ ! -f CHANGELOG.md ]; then
    echo "Error: CHANGELOG.md not found"
    return 1
  fi

  if ! grep -q "^## \[Unreleased\]" CHANGELOG.md; then
    echo "Error: CHANGELOG.md is missing [Unreleased] section"
    return 1
  fi

  local unreleased_content
  unreleased_content=$(sed -n '/^## \[Unreleased\]/,/^## \[/p' CHANGELOG.md | tail -n +2 | head -n -1)
  if [ -z "$(echo "$unreleased_content" | tr -d ' \n\t')" ]; then
    echo "Error: CHANGELOG.md [Unreleased] section is empty"
    return 1
  fi

  return 0
}

release_lib_validate_goreleaser() {
  if command -v goreleaser &> /dev/null; then
    if ! goreleaser check > /dev/null 2>&1; then
      echo "Error: GoReleaser configuration is invalid"
      echo "Run 'goreleaser check' to see errors"
      return 1
    fi
  fi
  return 0
}

release_lib_validate_commits_since_tag() {
  local latest_tag=$1

  if [ -n "$(git log --oneline "$latest_tag"..HEAD 2>/dev/null)" ]; then
    return 0
  else
    echo "Error: No commits since $latest_tag"
    return 1
  fi
}

release_lib_get_latest_tag() {
  git describe --tags --abbrev=0 2>/dev/null || echo ""
}

release_lib_validate_tag_format() {
  local tag=$1

  if [ -z "$tag" ]; then
    echo "Error: No tags found. Create initial tag with: git tag v0.1.0"
    return 1
  fi

  if [[ ! "$tag" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Invalid latest tag format: $tag (expected vX.Y.Z)"
    return 1
  fi

  return 0
}

release_lib_parse_version() {
  local tag=$1

  local version="${tag#v}"
  IFS='.' read -r release_lib_major release_lib_minor release_lib_patch <<< "$version"

  if ! [[ "$release_lib_major" =~ ^[0-9]+$ && "$release_lib_minor" =~ ^[0-9]+$ && "$release_lib_patch" =~ ^[0-9]+$ ]]; then
    echo "Error: Invalid version number format in tag: $tag"
    return 1
  fi

  echo "$release_lib_major $release_lib_minor $release_lib_patch"
}

release_lib_bump_version() {
  local bump_type=$1
  local major minor patch
  read -r major minor patch <<< "$(release_lib_parse_version "$2")"

  case $bump_type in
    patch)
      local new_patch=$((patch + 1))
      echo "${major}.${minor}.${new_patch}"
      ;;
    minor)
      local new_minor=$((minor + 1))
      echo "${major}.${new_minor}.0"
      ;;
    major)
      local new_major=$((major + 1))
      echo "${new_major}.0.0"
      ;;
    *)
      echo "Error: Invalid bump type '$bump_type'"
      return 1
      ;;
  esac
}

release_lib_check_tag_exists() {
  local tag=$1
  git tag -l "$tag" | grep -q "^${tag}$"
}

release_lib_find_release_commit() {
  local new_tag=$1
  git log --oneline -1 --all --grep "Release $new_tag" -- CHANGELOG.md 2>/dev/null | awk '{print $1}'
}

release_lib_generate_changelog_block() {
  local version=$1
  local date=$2
  local temp_file=$3

  awk -v ver="$version" -v dt="$date" '
    BEGIN { found = 0 }
    /^## \['"${version}"'\]/ {
      found = 1
      print
      while ((getline line) > 0) {
        if (line ~ /^## \[/) exit
        print line
      }
      exit
    }
    END { exit !found }
  ' "$temp_file"
}

release_lib_update_changelog() {
  local new_version=$1
  local temp_file=$2

  local today
  today=$(date +%Y-%m-%d)

  awk -v version="$new_version" -v date="$today" '
    /^## \[Unreleased\]/ {
      print "## [" version "] - " date
      print ""
      in_unreleased = 1
      next
    }
    in_unreleased && /^## \[/ {
      in_unreleased = 0
      print
      next
    }
    in_unreleased { print; next }
    { print }
  ' CHANGELOG.md > "$temp_file"
}

release_lib_check() {
  local bump_type=$1

  echo "=== Validating release prerequisites ==="
  echo ""

  echo "[1/4] Validating environment..."
  release_lib_validate_environment
  echo "✅ Environment OK (git repo, main branch, clean working directory)"
  echo ""

  echo "[2/4] Validating CHANGELOG.md..."
  release_lib_validate_changelog
  echo "✅ CHANGELOG.md OK ([Unreleased] section present with content)"
  echo ""

  echo "[3/4] Validating GoReleaser..."
  release_lib_validate_goreleaser
  echo "✅ GoReleaser OK (or not installed)"
  echo ""

  LATEST_TAG=$(release_lib_get_latest_tag)
  release_lib_validate_tag_format "$LATEST_TAG"
  release_lib_validate_commits_since_tag "$LATEST_TAG"

  NEW_VERSION=$(release_lib_bump_version "$bump_type" "$LATEST_TAG")
  NEW_TAG="v${NEW_VERSION}"

  echo "[4/4] Version bump calculation..."
  echo "✅ Current tag: $LATEST_TAG"
  echo "✅ New tag would be: $NEW_TAG"
  echo ""

  if release_lib_check_tag_exists "$NEW_TAG"; then
    echo "⚠️  Tag $NEW_TAG already exists - would skip commit creation"
  else
    echo "✅ Tag $NEW_TAG does not exist - would create commit"
  fi

  echo ""
  echo "✅ All checks passed. Run 'mise run release $bump_type' to execute."
}

#!/usr/bin/env bash

# some snippets:
#
# gum confirm "Commit changes?" && git commit -m "$SUMMARY" -m "$DESCRIPTION"

# 🍭 gum utility
#

#
# 🍭 end gum utility

# 🍯 git dev workflow commands; This script makes use of nerdfonts.com glyphs, eg 
#

# === 🥥 git-operations ========================================================

function get-def-branch() {
  echo master
}

function gad() {
  if [ -z "$(git status -s -uno | grep -v '^ ' | awk '{print $2}')" ]; then
      gum confirm "Stage all?" && git add .
  fi

  return 0
}

function get-tracking-branch() {
  git config "branch.$(git_current_branch).remote"
}

function check-upstream-branch() {
  upstream_branch=$(get-tracking-branch)
  feature_branch=$(git_current_branch)

  if [ -z "$upstream_branch" ]; then
    echo "===> 🐛 No upstream branch detected for : '🎀 $feature_branch'"

    if ! are-you-sure; then
      return 1
    fi
  fi

  return 0
}

# === 🥥 prompt ================================================================

function _prompt() {
  # gum confirm exits with status 0 if confirmed and status 1 if cancelled
  # message="$1"
  # result=$(gum confirm "$message")

  # return "$result"

  message="$1"
  gum confirm "$message"

  return $?
}

function _prompt-are-you-sure {
  _prompt "are you sure? 👾"
  result=$?

  if [ ! "$result" -eq 0 ]; then
    echo "⛔ Aborted!"
  fi

  return $result
}

# this may no longer be required, becaause the gum confirm can be
# integrated into the call site; or define a prompt_ wrapper function
#
are-you-sure() { # 
  echo "👾 Are you sure❓ (type 'y' to confirm)"
  # $(gum input --placeholder "scope")

  # squashed=$(gum input --placeholder "scope")

  read -r squashed

  if [ ! "$squashed" = "y" ]; then
    echo "⛔ Aborted!"

    return 1
  fi

  return 0
}

# === 🥥 start-feat ============================================================

function start-feat() {
  if [[ -n $1 ]]; then
    echo "===> 🚀 START FEATURE: '🎀 $1'"

    git checkout -b "$1"
  else
    echo "!!! 😕 Please specify a feature branch"

    return 1
  fi

  return 0
}

# === 🥥 end-feat ==============================================================

function _do-end-feat() {
  feature_branch=$(git_current_branch)
  default_branch=$(get-def-branch)

  if _prompt "About to end feature 🎁 '$feature_branch' ... have you squashed commits"; then
    echo "<=== ✨ END FEATURE: '$feature_branch'"

    if [ "$feature_branch" != master ] && [ "$feature_branch" != main ]; then
      git branch --unset-upstream
      # can't reliably use prune here, because we have in effect
      # a race condition, depending on how quickly github deletes
      # the upstream branch after Pull Request "Rebase and Merge"
      #
      # gcm && git fetch --prune
      # have to wait a while and perform the prune or delete
      # local branch manually.
      #
      git checkout "$default_branch"
      git pull origin "$default_branch"
      echo "Done! ✅"
    else
      echo "!!! 😕 Not on a feature branch ($feature_branch)"

      return 1
    fi
  else
    echo "⛔ Aborted!"

    return 1
  fi

  return 0
}

function end-feat() {
  _prompt-are-you-sure && _do-end-feat
}

# === 🥥 push-feat =============================================================

function _do-push-feat() {
  current_branch=$(git_current_branch)
  default_branch=$(get-def-branch)

  if [ "$current_branch" = "$default_branch" ]; then
    echo "!!! ⛔ Aborted! still on default branch($default_branch) branch ($current_branch)"

    return 1
  fi

  if ! git push origin --set-upstream "$current_branch"; then
    echo "!!! ⛔ Aborted! Failed to push feature for branch: $current_branch"

    return 1
  fi

  echo "pushed feature to $current_branch, ok! ✅"

  return 0
}

function push-feat() {
  _prompt-are-you-sure && _do-push-feat
}

# === 🥥 check-tag =============================================================

function check-tag() {
  rel_tag=$1
  if ! [[ $rel_tag =~ ^[0-9] ]]; then
    echo "!!! ⛔ Aborted! invalid tag"

    return 1
  fi

  return 0
}

# === 🥥 do-release ============================================================

function _do-release() {
  if [[ -n $1 ]]; then
    if ! check-tag "$1"; then
      return 1
    fi

    raw_version=$1
    version_number=v$1
    current_branch=$(git_current_branch)
    default_branch=$(get-def-branch)

    if [[ $raw_version == v* ]]; then
      # the # in ${raw_version#v} is a parameter expansion operator
      # that removes the shortest match of the pattern "v" from the beginning
      # of the string variable.
      #
      version_number=$raw_version
      raw_version=${raw_version#v}
    fi

    if [ "$current_branch" != "$default_branch" ]; then
      echo "!!! ⛔ Aborted! not on default($default_branch) branch; current($current_branch)"

      return 1
    fi

    echo "===> 🚀 START RELEASE: '🎁 $version_number'"
    release_branch=release/$version_number

    if ! git checkout -b "$release_branch"; then
      echo "!!! ⛔ Aborted! Failed to create the release branch: $release_branch"

      return 1
    fi

    if [ -e ./VERSION ]; then      
      if ! echo "$version_number" > ./VERSION; then
        echo "!!! ⛔ Aborted! Failed to update VERSION file"

        return 1
      fi

      
      if ! git add ./VERSION; then
        echo "!!! ⛔ Aborted! Failed to git add VERSION file"

        return 1
      fi

      if [ -e ./VERSION-PATH ]; then
        version_path=$(more ./VERSION-PATH)
        echo "$raw_version" > "$version_path"
        
        if ! git add "$version_path"; then
          echo "!!! ⛔ Aborted! Failed to git add VERSION-PATH file ($version_path)"

          return 1
        fi
      fi

      if ! git commit -m "Bump version to $version_number"; then
        echo "!!! ⛔ Aborted! Failed to commit VERSION file"

        return 1
      fi
      
      if ! git push origin --set-upstream "$release_branch"; then
        echo "!!! ⛔ Aborted! Failed to push release $version_number upstream"

        return 1
      fi

      echo "Done! ✅"
    else
      echo "!!! ⛔ Aborted! VERSION file is missing. (In root dir?)"

      return 1
    fi
  else
    echo "!!! 😕 Please specify a semantic version to release"

    return 1
  fi

  return 0
}

# release <semantic-version>, !!! do not specify the v prefix, added automatically
# should be run from the root directory otherwise relative paths won't work properly.
function release() {
  _prompt-are-you-sure && _do-release "$1"
}

# === 🥥 tag-rel ===============================================================

function _do-tag-rel() {
  if [[ -n "$1" ]]; then
    version_number="v$1"
    current_branch=$(git_current_branch)
    default_branch=$(get-def-branch)

    if [ "$current_branch" != "$default_branch" ]; then
      echo "!!! ⛔ Aborted! not on default($default_branch) branch; current($current_branch)"

      return 1
    fi

    echo "===> 🏷️ PUSH TAG: '$version_number'"

    
    if ! git tag -a "$version_number" -m "Release $version_number"; then
      echo "!!! ⛔ Aborted! Failed to create annotated tag: $version_number"

      return 1
    fi

    
    if ! git push origin "$version_number"; then
      echo "!!! ⛔ Aborted! Failed to push tag: $version_number"

      return 1
    fi

    echo "Done! ✅"
  else
    echo "!!! 😕 Please specify a release semantic version to tag"

    reurn 1
  fi

  return 0
}

# tag-rel <semantic-version>, !!! do not specify the v prefix, added automatically
function tag-rel() {
  _prompt-are-you-sure && _do-tag-rel "$1"
}

#
# 🍯 end git dev workflow commands

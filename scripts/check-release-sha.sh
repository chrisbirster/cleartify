#!/usr/bin/env bash
# Offline integrity check for a full release/CI source commit SHA.
set -euo pipefail
if [[ "$#" -ne 1 ]]; then echo "usage: check-release-sha.sh <40-hex-sha>" >&2; exit 2; fi
sha="$1"
if [[ ! "$sha" =~ ^[0-9a-fA-F]{40}$ ]]; then
  echo "invalid release SHA: expected a full 40-character hexadecimal commit ID" >&2
  exit 2
fi
if ! git cat-file -e "${sha}^{commit}" 2>/dev/null; then
  echo "release SHA not present as a commit in local checkout" >&2
  exit 1
fi
head="$(git rev-parse HEAD)"
resolved="$(git rev-parse "${sha}^{commit}")"
if [[ "$head" != "$resolved" ]]; then
  echo "release SHA does not match checked-out HEAD" >&2
  exit 1
fi
printf 'validated release source commit: %s\n' "$head"
